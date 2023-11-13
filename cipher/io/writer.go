// Copyright 2023 The Bumblebee Authors
//
// Use of this source code is governed by an MIT license that is located
// in this project's root folder, and can also be found online at:
//
// https://github.com/thoughtrealm/bumblebee/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package io

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CipherFileWriterIntf interface {
}

type CipherWriter struct {
	ReceiverCipherPublicKey string
	SenderCipherKeyPair     nkeys.KeyPair
	SenderSigningKeyPair    nkeys.KeyPair
	CombinedFilePath        string
	BundleFilePath          string
	DataFilePath            string
	SymmetricKey            []byte
	OutputBundleInfo        *BundleInfo
	SymmetricCipher         beecipher.Cipher
}

func NewCipherWriter(receiverKI *security.KeyInfo, senderKPI *security.KeyPairInfo) (*CipherWriter, error) {
	if receiverKI == nil {
		return nil, errors.New("receiver key is nil")
	}

	if senderKPI == nil {
		return nil, errors.New("sender key is nil")
	}

	SenderCipherKeyPair, err := nkeys.FromCurveSeed(senderKPI.CipherSeed)
	if err != nil {
		return nil, fmt.Errorf("error transforming sender key seed: %w", err)
	}

	SenderSigningKP, err := nkeys.FromSeed(senderKPI.SigningSeed)
	if err != nil {
		return nil, fmt.Errorf("error transforming sender key seed: %w", err)
	}

	bundleInfo, err := NewBundle()
	if err != nil {
		return nil, fmt.Errorf("failed generating bundle header: %w", err)
	}

	bundleInfo.FromName = senderKPI.Name
	bundleInfo.ToName = receiverKI.Name
	bundleInfo.SenderSig, err = senderKPI.SignRandom()
	if err != nil {
		return nil, fmt.Errorf("failed generating bundle header: %w", err)
	}

	return &CipherWriter{
		SenderCipherKeyPair:     SenderCipherKeyPair,
		SenderSigningKeyPair:    SenderSigningKP,
		ReceiverCipherPublicKey: receiverKI.CipherPubKey,
		OutputBundleInfo:        bundleInfo,
	}, nil
}

func (cfw *CipherWriter) WriteToCombinedFileFromReader(combinedFilePath string, r io.Reader) (int, error) {
	usePath := combinedFilePath
	ext := strings.ToLower(filepath.Ext(usePath))
	if ext == "" || ext == ".ext" {
		usePath = helpers.ReplaceFileExt(combinedFilePath, ".bcomb")
	}

	bcombFile, err := os.Create(usePath)
	if err != nil {
		return 0, fmt.Errorf("failed creating bcomb output file: %s", err)
	}

	defer func() {
		_ = bcombFile.Close()
	}()

	headerBytesWritten, err := cfw.WriteBundleHeader(bcombFile)
	if err != nil {
		return headerBytesWritten, fmt.Errorf("unable to write bundle header: %w", err)
	}

	dataBytesWritten, err := cfw.WriteBundleData(r, bcombFile)
	if err != nil {
		return dataBytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return headerBytesWritten + dataBytesWritten, nil
}

func (cfw *CipherWriter) WriteToSplitFilesFromReader(combinedFilePath string, r io.Reader) (int, error) {
	bhdrFilePath := helpers.ReplaceFileExt(combinedFilePath, ".bhdr")
	bhdrFile, err := os.Create(bhdrFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed creating bhdr output file: %s", err)
	}

	defer func() {
		_ = bhdrFile.Close()
	}()

	bdataFilePath := helpers.ReplaceFileExt(combinedFilePath, ".bdata")
	bdataFile, err := os.Create(bdataFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed creating bdata output file: %s", err)
	}

	defer func() {
		_ = bdataFile.Close()
	}()

	headerBytesWritten, err := cfw.WriteBundleHeader(bhdrFile)
	if err != nil {
		return headerBytesWritten, fmt.Errorf("unable to write bundle header: %w", err)
	}

	dataBytesWritten, err := cfw.WriteBundleData(r, bdataFile)
	if err != nil {
		return dataBytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return headerBytesWritten + dataBytesWritten, nil
}

type StreamCompleteFunc func(w io.Writer) error

func (cfw *CipherWriter) WriteToCombinedStreamFromReader(r io.Reader, w io.Writer, completeFunc StreamCompleteFunc) (int, error) {
	headerBytesWritten, err := cfw.WriteBundleHeader(w)
	if err != nil {
		return headerBytesWritten, fmt.Errorf("unable to write bundle header: %w", err)
	}

	dataBytesWritten, err := cfw.WriteBundleData(r, w)
	if err != nil {
		return headerBytesWritten + dataBytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	if completeFunc != nil {
		return headerBytesWritten + dataBytesWritten, completeFunc(w)
	}

	return headerBytesWritten + dataBytesWritten, nil
}

func (cfw *CipherWriter) WriteToSplitStreamsFromReader(r io.Reader, wHdr io.Writer, wData io.Writer, completeFuncHdr, completeFuncData StreamCompleteFunc) (int, error) {
	headerBytesWritten, err := cfw.WriteBundleHeader(wHdr)
	if err != nil {
		return headerBytesWritten, fmt.Errorf("unable to write bundle header: %w", err)
	}

	if completeFuncHdr != nil {
		err = completeFuncHdr(wHdr)
		if err != nil {
			return headerBytesWritten, err
		}
	}

	dataBytesWritten, err := cfw.WriteBundleData(r, wData)
	if err != nil {
		return headerBytesWritten + dataBytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	if completeFuncData != nil {
		return headerBytesWritten + dataBytesWritten, completeFuncData(wData)
	}

	return headerBytesWritten + dataBytesWritten, nil
}

func (cfw *CipherWriter) WriteBundleHeader(writer io.Writer) (int, error) {
	var err error
	cfw.SymmetricCipher, err = beecipher.NewSymmetricCipher(cfw.OutputBundleInfo.SymmetricKey, CHUNK_SIZE)
	if err != nil {
		return 0, fmt.Errorf("failed generating symmetric sc: %s", err)
	}

	cfw.OutputBundleInfo.Salt = cfw.SymmetricCipher.GetSalt()

	bundleBytes, err := msgpack.Marshal(cfw.OutputBundleInfo)
	if err != nil {
		return 0, fmt.Errorf("failed serializing bundle info: %s", err)
	}
	defer cfw.OutputBundleInfo.Wipe()
	defer security.Wipe(bundleBytes)

	senderCipherSeed, err := cfw.SenderCipherKeyPair.Seed()
	if err != nil {
		return 0, fmt.Errorf("failed extracting sender cipher seed: %s", err)
	}
	defer security.Wipe(senderCipherSeed)

	bundleWriterBuff := bytes.NewBuffer(nil)
	bundleReaderBuff := bytes.NewBuffer(bundleBytes)
	nc, err := beecipher.NewNKeysCipherEncrypter(cfw.ReceiverCipherPublicKey, senderCipherSeed)
	if err != nil {
		return 0, fmt.Errorf("failed creating new nkeys sc encrypter: %s", err)
	}
	defer nc.Wipe()

	_, err = nc.Encrypt(bundleReaderBuff, bundleWriterBuff)
	if err != nil {
		return 0, fmt.Errorf("failed encrypting bundle info: %s", err)
	}

	encryptedBundleBytes := bundleWriterBuff.Bytes()
	bundleLenBytes := IntToUint16Bytes(len(encryptedBundleBytes))

	// First, write out the bundle header len
	lenBytesWritten, err := writer.Write(bundleLenBytes)
	if err != nil {
		return lenBytesWritten, fmt.Errorf("failed writing bundle length: %s", err)
	}
	if lenBytesWritten != len(bundleLenBytes) {
		return lenBytesWritten, fmt.Errorf(
			"error bundle length: wrote %d bytes, expected %d bytes",
			lenBytesWritten,
			len(bundleLenBytes),
		)
	}

	// Now, write out the bundle header
	headerBytesWritten, err := writer.Write(encryptedBundleBytes)
	if err != nil {
		return lenBytesWritten, fmt.Errorf("failed writing encrypted bundle data: %s", err)
	}
	if headerBytesWritten != len(encryptedBundleBytes) {
		return headerBytesWritten, fmt.Errorf(
			"error bundle length: wrote %d bytes, expected %d bytes",
			headerBytesWritten,
			len(encryptedBundleBytes),
		)
	}

	return lenBytesWritten + headerBytesWritten, nil
}

func (cfw *CipherWriter) WriteBundleData(r io.Reader, w io.Writer) (int, error) {
	// encode the data stream to the file
	return cfw.SymmetricCipher.Encrypt(r, w)
}

type LenMarkerSize int

const (
	LenMarkerSize8 LenMarkerSize = iota
	LenMarkerSize16
	LenMarkerSize32
	LenMarkerSize64
)

func WriteBytesTo(data []byte, markerSize LenMarkerSize, w io.Writer) (n int, err error) {
	var bytesWritten int
	switch markerSize {
	case LenMarkerSize8:
		bytesWritten, err = WriteUint8Marker(len(data), w)
	case LenMarkerSize16:
		bytesWritten, err = WriteUint16Marker(len(data), w)
	default:
		return bytesWritten, fmt.Errorf("Length Marker Not Supported: %d", int(markerSize))
	}

	if err != nil {
		return bytesWritten, fmt.Errorf("Error writing marker: %s", err)
	}

	n, err = w.Write(data)
	bytesWritten += n
	if err != nil {
		return bytesWritten, fmt.Errorf("error writing bytes data: %s", err)
	}
	if n != len(data) {
		return bytesWritten, fmt.Errorf(
			"error writing bytes data: write count wrong: wrote %d bytes, expectd %d",
			n,
			len(data),
		)
	}

	return bytesWritten, nil
}

func WriteUint8Marker(val int, w io.Writer) (n int, err error) {
	uint8Bytes := IntToUint8Bytes(val)
	n, err = w.Write(uint8Bytes)
	if err != nil {
		return n, err
	}
	if n != len(uint8Bytes) {
		return n, fmt.Errorf("write count wrong: %d bytes written, expected %d", n, len(uint8Bytes))
	}

	return n, nil
}

func WriteUint16Marker(val int, w io.Writer) (n int, err error) {
	uint16Bytes := IntToUint16Bytes(val)
	n, err = w.Write(uint16Bytes)
	if err != nil {
		return n, err
	}
	if n != len(uint16Bytes) {
		return n, fmt.Errorf("write count wrong: %d bytes written, expected %d", n, len(uint16Bytes))
	}

	return n, nil
}

func (cfw *CipherWriter) Wipe() {
	if cfw.SenderCipherKeyPair != nil {
		cfw.SenderCipherKeyPair.Wipe()
	}

	if cfw.SenderSigningKeyPair != nil {
		cfw.SenderSigningKeyPair.Wipe()
	}
}
