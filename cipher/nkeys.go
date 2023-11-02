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

package cipher

import (
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"io"
)

type NKeysCipher struct {
	ReceiverKP     nkeys.KeyPair
	ReceiverPubKey string
	SenderKP       nkeys.KeyPair
	SenderPubKey   string
	BytesRead      int
	BytesWritten   int
}

func NewNKeysCipherDecrypter(receiverSeed []byte, senderPubKey string) (*NKeysCipher, error) {
	if receiverSeed == nil {
		return nil, errors.New("receiverSeed is empty")
	}

	if senderPubKey == "" {
		return nil, errors.New("senderPubKey is empty")
	}

	newNKeysCipher := &NKeysCipher{}
	var err error

	newNKeysCipher.ReceiverKP, err = nkeys.FromSeed(receiverSeed)
	if err != nil {
		return nil, fmt.Errorf("failed creating receiver keypair: %w", err)
	}
	newNKeysCipher.SenderPubKey = senderPubKey

	return newNKeysCipher, nil
}

func NewNKeysCipherEncrypter(receiverPubKey string, senderSeed []byte) (*NKeysCipher, error) {
	if receiverPubKey == "" {
		return nil, errors.New("receiverPubKey is empty")
	}

	if senderSeed == nil {
		return nil, errors.New("senderSeed is empty")
	}

	newNKeysCipher := &NKeysCipher{}
	var err error

	newNKeysCipher.SenderKP, err = nkeys.FromSeed(senderSeed)
	if err != nil {
		return nil, fmt.Errorf("failed creating sender keypair: %w", err)
	}
	newNKeysCipher.ReceiverPubKey = receiverPubKey

	return newNKeysCipher, nil
}

func (nk *NKeysCipher) GetBytesRead() int {
	return nk.BytesRead
}

func (nk *NKeysCipher) GetBytesWritten() int {
	return nk.BytesWritten
}

func (nk *NKeysCipher) GetChunkSize() int {
	return 0
}

func (nk *NKeysCipher) GetDerivedKey() []byte {
	return nil
}

func (nk *NKeysCipher) GetSalt() []byte {
	return nil
}

func (nk *NKeysCipher) Decrypt(r io.Reader, w io.Writer) (int, error) {
	// First, read in the bytes to decrypt
	var encryptedBytes []byte
	encryptedBuff := make([]byte, 16000)
	for {
		bytesRead, err := r.Read(encryptedBuff)
		if bytesRead > 0 {
			nk.BytesRead += bytesRead
			encryptedBytes = append(encryptedBytes, encryptedBuff[:bytesRead]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("error reading input stream: %w", err)
		}
	}

	decryptedBytes, err := nk.ReceiverKP.Open(encryptedBytes, nk.SenderPubKey)
	if err != nil {
		return 0, fmt.Errorf("failed decrypting input: %w", err)
	}

	bytesWritten, err := w.Write(decryptedBytes)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed sending encrypted data to output: %w", err)
	}

	nk.BytesWritten = bytesWritten
	if bytesWritten != len(decryptedBytes) {
		return bytesWritten, fmt.Errorf(
			"failed sending data to output: sent %d bytes, expected to send %d bytes",
			bytesWritten,
			len(decryptedBytes),
		)
	}

	return bytesWritten, nil
}

func (nk *NKeysCipher) Encrypt(r io.Reader, w io.Writer) (int, error) {
	// First, read in the bytes to encrypt
	var inputBytes []byte
	inputBuff := make([]byte, 16000)
	for {
		bytesRead, err := r.Read(inputBuff)
		if bytesRead > 0 {
			nk.BytesRead += bytesRead
			inputBytes = append(inputBytes, inputBuff[:bytesRead]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("error reading input stream: %w", err)
		}
	}

	encryptedBytes, err := nk.SenderKP.Seal(inputBytes, nk.ReceiverPubKey)
	if err != nil {
		return 0, fmt.Errorf("failed encrypting input: %w", err)
	}

	bytesWritten, err := w.Write(encryptedBytes)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed sending encrypted data to output: %w", err)
	}

	nk.BytesWritten = bytesWritten
	if bytesWritten != len(encryptedBytes) {
		return bytesWritten, fmt.Errorf(
			"failed sending data to output: sent %d bytes, expected to send %d bytes",
			bytesWritten,
			len(encryptedBytes),
		)
	}

	return bytesWritten, nil
}
