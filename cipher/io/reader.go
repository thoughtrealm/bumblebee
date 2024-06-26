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
	"github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/thoughtrealm/bumblebee/streams"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"os"
	"path/filepath"
)

const DEFAULT_OUTPUT_FILE_NAME = "bee.output"

type CipherReader struct {
	ReceiverCipherKP    nkeys.KeyPair
	SenderCipherPubKey  string
	SenderSigningPubKey string
	CombinedFilePath    string
	BundleFilePath      string
	DataFilePath        string
}

func NewCipherFileReader(receiverKPI *security.KeyPairInfo, senderKI *security.KeyInfo) (*CipherReader, error) {
	if receiverKPI == nil {
		return nil, errors.New("receiver key is nil")
	}

	if senderKI == nil {
		return nil, errors.New("sender key is nil")
	}

	ReceiverKP, err := receiverKPI.GetCipherKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error transforming receiver cipher seed: %w", err)
	}

	return &CipherReader{
		ReceiverCipherKP:    ReceiverKP,
		SenderCipherPubKey:  senderKI.CipherPubKey,
		SenderSigningPubKey: senderKI.SigningPubKey,
	}, nil
}

func (cfr *CipherReader) ReadCombinedFileToBytes(combinedFilePath string) ([]byte, error) {
	fileIn, err := os.Open(combinedFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed opening combined file: %w", err)
	}

	defer func() {
		_ = fileIn.Close()
	}()

	// For file writing, we have to read the bundle first, because it contains the target filename.
	bundleInfo, err := cfr.readBundleHeaderFrom(fileIn, false)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bundle from input: %w", err)
	}
	defer bundleInfo.Wipe()

	outputBuffer := bytes.NewBuffer(nil)
	_, err = cfr.readBundleDataTo(bundleInfo, fileIn, outputBuffer)
	if err != nil {
		return nil, fmt.Errorf("unable to write bundle body: %w", err)
	}

	return outputBuffer.Bytes(), nil
}

func (cfr *CipherReader) GetBundleDetailsFromReader(r io.Reader) (*BundleInfo, error) {
	return cfr.readBundleHeaderFrom(r, true)
}

func (cfr *CipherReader) readBundleHeaderFrom(r io.Reader, allowMultiDir bool) (*BundleInfo, error) {
	// Get the bundle len first
	bundleLenBytes := make([]byte, 2)
	bytesRead, err := r.Read(bundleLenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed reading bundle length from input: %w", err)
	}
	if bytesRead != len(bundleLenBytes) {
		return nil, fmt.Errorf("failed reading bundle len data from input: read %d bytes, expected %d bytes",
			bytesRead,
			len(bundleLenBytes),
		)
	}

	bundleLen, err := Uint16BytesToInt(bundleLenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed transforming bundle len data: %w", err)
	}

	// Now, read in the encrypted bundle info
	encryptedBundleBytes := make([]byte, bundleLen)
	bytesRead, err = r.Read(encryptedBundleBytes)
	if err != nil {
		return nil, fmt.Errorf("failed reading bundle data from input: %w", err)
	}
	if bytesRead != bundleLen {
		return nil, fmt.Errorf("failed reading bundle data from input: read %d bytes, expected %d bytes",
			bytesRead,
			bundleLen,
		)
	}

	receiverSeed, err := cfr.ReceiverCipherKP.Seed()
	if err != nil {
		return nil, fmt.Errorf("failed extracting seed from receiver kp: %w", err)
	}
	defer security.Wipe(receiverSeed)

	nc, err := cipher.NewNKeysCipherDecrypter(receiverSeed, cfr.SenderCipherPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed creating nkeys cipher: %w", err)
	}
	defer nc.Wipe()

	bundleDecryptReader := bytes.NewBuffer(encryptedBundleBytes)
	bundleDecryptWriter := bytes.NewBuffer(nil)

	_, err = nc.Decrypt(bundleDecryptReader, bundleDecryptWriter)
	if err != nil {
		return nil, fmt.Errorf("failed decrypting bundle header: %w", err)
	}

	bundleDecrytedBytes := bundleDecryptWriter.Bytes()
	bundleInfo := &BundleInfo{}
	err = msgpack.Unmarshal(bundleDecrytedBytes, bundleInfo)
	if err != nil {
		return nil, fmt.Errorf("failed transforming bundle header: %w", err)
	}

	logger.Debug("Validating bundle signature")
	verifyKI, _ := security.NewKeyInfo("verify-sender", cfr.SenderCipherPubKey, cfr.SenderSigningPubKey)
	isValid, err := verifyKI.VerifyRandomSignature(bundleInfo.SenderSig)
	if err != nil {
		logger.Debugfln("Sender identity validation failed: %s", err)
		return nil, fmt.Errorf("Sender identity validation failed: %w", err)
	}

	if !isValid {
		// Todo: Validate sender sig... we may want a warning flag so you can override this hard error
		logger.Debug("Bundle signature does not match sender identity")
		return nil, errors.New("bundle signature does not match sender identity")
	}

	if !allowMultiDir && bundleInfo.InputSource == BundleInputSourceMultiDir {
		return nil, errors.New("this bundle access request does not support multi-directory bundles")
	}

	return bundleInfo, nil
}

func (cfr *CipherReader) readBundleDataTo(bundleInfo *BundleInfo, r io.Reader, w io.Writer) (int, error) {
	sc, err := cipher.NewSymmetricCipherFromSalt(bundleInfo.SymmetricKey, bundleInfo.Salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return 0, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	bytesWritten, err := sc.Decrypt(r, w)
	return bytesWritten, err
}

// ReadCombinedFileToWriter assumes the input file path provided has been validated
func (cfr *CipherReader) ReadCombinedFileToWriter(combinedFilePath string, w io.Writer) (int, error) {
	fileIn, err := os.Open(combinedFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening combined file: %w", err)
	}

	defer func() {
		_ = fileIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileIn, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileIn, w)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

// ReadCombinedFileToPath assumes the input file path provided has been validated
func (cfr *CipherReader) ReadCombinedFileToPath(combinedFilePath, outputPath string) (int, error) {
	fileIn, err := os.Open(combinedFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening combined file: %w", err)
	}

	defer func() {
		_ = fileIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileIn, true)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	var outputWriter io.Writer
	var mdsw streams.StreamWriter
	if bundleInfo.InputSource == BundleInputSourceMultiDir {
		// For multi dir streams, we just need to initialize the stream writer with the output path
		// Todo: Maybe add support for the includePaths from symfile support?
		mdsw, err = streams.NewMultiDirectoryStreamWriter(outputPath, false, nil)
		if err != nil {
			return 0, fmt.Errorf("unable to initialize multi directory stream writer: %w", err)
		}

		outputWriter, err = mdsw.StartStream()
		if err != nil {
			return 0, fmt.Errorf("unable to start multi directory stream: %w", err)
		}
	} else {
		var outputFilePath string
		if bundleInfo.OriginalFileName == "" {
			// get input filename
			_, fileName := filepath.Split(combinedFilePath)
			outputFilePath = filepath.Join(outputPath, helpers.ReplaceFileExt(fileName, ".decrypted"))
		} else {
			outputFilePath = filepath.Join(outputPath, bundleInfo.OriginalFileName)
		}

		fileOut, err := os.Create(outputFilePath)
		if err != nil {
			return 0, fmt.Errorf("unable to open output file: %w", err)
		}

		defer func() {
			_ = fileOut.Close()
		}()

		outputWriter = fileOut
	}

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileIn, outputWriter)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	if mdsw != nil {
		// The multi dir writer may emit more data than the decryptor is aware of.
		// So, get the total from the multi dir writer when one is used.
		return mdsw.TotalBytesWritten(), nil
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) ReadStreamToPath(reader io.Reader, outputPath string) (int, error) {
	bundleInfo, err := cfr.readBundleHeaderFrom(reader, true)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	var outputFilePath string
	if bundleInfo.OriginalFileName == "" {
		outputFilePath = filepath.Join(outputPath, DEFAULT_OUTPUT_FILE_NAME+".decrypted")
	} else {
		outputFilePath = filepath.Join(outputPath, bundleInfo.OriginalFileName)
	}

	fileOut, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open output file: %w", err)
	}

	defer func() {
		_ = fileOut.Close()
	}()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, reader, fileOut)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

// ReadCombinedFileToFile assumes the input file path provided has been validated.
// This func does not consider the filename in the header.  It just writes the output to the provided outputFilePath.
func (cfr *CipherReader) ReadCombinedFileToFile(combinedFilePath, outputFilePath string) (int, error) {
	fileIn, err := os.Open(combinedFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening combined file: %w", err)
	}

	defer func() {
		_ = fileIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileIn, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	fileOut, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open output file: %w", err)
	}

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileIn, fileOut)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) ReadStreamToFile(reader io.Reader, outputFilePath string) (int, error) {
	bundleInfo, err := cfr.readBundleHeaderFrom(reader, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	fileOut, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open output file: %s", err)
	}

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, reader, fileOut)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

// ReadSplitFilesToWriter assumes the input file path provided has been validated
func (cfr *CipherReader) ReadSplitFilesToWriter(bundleHeaderFilePath, bundleDataFilePath string, w io.Writer) (int, error) {
	fileHdrIn, err := os.Open(bundleHeaderFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileHdrIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileHdrIn, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	fileDataIn, err := os.Open(bundleDataFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileDataIn.Close()
	}()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileDataIn, w)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

// ReadSplitFilesToPath assumes the input file paths provided have been validated
func (cfr *CipherReader) ReadSplitFilesToPath(bundleHeaderFilePath, bundleDataFilePath, outputPath string) (int, error) {
	fileHdrIn, err := os.Open(bundleHeaderFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileHdrIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileHdrIn, true)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	var outputWriter io.Writer
	var mdsw streams.StreamWriter
	if bundleInfo.InputSource == BundleInputSourceMultiDir {
		// For multi dir streams, we just need to initialize the stream writer with the output path
		// Todo: Maybe add support for the includePaths from symfile support?
		mdsw, err = streams.NewMultiDirectoryStreamWriter(outputPath, false, nil)
		if err != nil {
			return 0, fmt.Errorf("unable to initialize multi directory stream writer: %w", err)
		}

		outputWriter, err = mdsw.StartStream()
		if err != nil {
			return 0, fmt.Errorf("unable to start multi directory stream: %w", err)
		}
	} else {
		var outputFilePath string
		if bundleInfo.OriginalFileName == "" {
			// get input filename
			_, fileName := filepath.Split(bundleHeaderFilePath)
			outputFilePath = filepath.Join(outputPath, helpers.ReplaceFileExt(fileName, ".decrypted"))
		} else {
			outputFilePath = filepath.Join(outputPath, bundleInfo.OriginalFileName)
		}

		fileOut, err := os.Create(outputFilePath)
		if err != nil {
			return 0, fmt.Errorf("unable to open output file: %s", err)
		}

		defer func() {
			_ = fileOut.Close()
		}()

		outputWriter = fileOut
	}

	fileDataIn, err := os.Open(bundleDataFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileDataIn.Close()
	}()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileDataIn, outputWriter)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	if mdsw != nil {
		// The multi dir writer may emit more data than the decryptor is aware of.
		// So, get the total from the multi dir writer when one is used.
		return mdsw.TotalBytesWritten(), nil
	}

	return bytesWritten, nil
}

// ReadSplitFilesToPath assumes the input file paths provided have been validated
func (cfr *CipherReader) ReadSplitFilesToFile(bundleHeaderFilePath, bundleDataFilePath, outputFilePath string) (int, error) {
	fileHdrIn, err := os.Open(bundleHeaderFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileHdrIn.Close()
	}()

	bundleInfo, err := cfr.readBundleHeaderFrom(fileHdrIn, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle header from input: %w", err)
	}
	defer bundleInfo.Wipe()

	fileOut, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open output file: %s", err)
	}

	defer func() {
		_ = fileOut.Close()
	}()

	fileDataIn, err := os.Open(bundleDataFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening bundle header file: %s", err)
	}

	defer func() {
		_ = fileDataIn.Close()
	}()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, fileDataIn, fileOut)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) ReadCombinedStreamToWriter(r io.Reader, w io.Writer) (int, error) {
	bundleInfo, err := cfr.readBundleHeaderFrom(r, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle from input: %w", err)
	}
	defer bundleInfo.Wipe()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, r, w)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) ReadSplitStreamsToWriter(readerHdr io.Reader, readerData io.Reader, w io.Writer) (int, error) {
	bundleInfo, err := cfr.readBundleHeaderFrom(readerHdr, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle hdr from input: %w", err)
	}
	defer bundleInfo.Wipe()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, readerData, w)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) ReadCombinedStreamToFile(r io.Reader, outputFilePath string) (int, error) {
	bundleInfo, err := cfr.readBundleHeaderFrom(r, false)
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve bundle from input: %w", err)
	}
	defer bundleInfo.Wipe()

	bytesWritten, err := cfr.readBundleDataTo(bundleInfo, r, nil)
	if err != nil {
		return bytesWritten, fmt.Errorf("unable to write bundle data: %w", err)
	}

	return bytesWritten, nil
}

func (cfr *CipherReader) Wipe() {
	if cfr.ReceiverCipherKP != nil {
		cfr.ReceiverCipherKP.Wipe()
	}
}
