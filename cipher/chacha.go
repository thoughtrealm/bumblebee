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
	"crypto/cipher"
	cryptorand "crypto/rand"
	"fmt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"strconv"
)

/*
	This implementation is based partly on a pattern demonstrated in the following article:
		Author: Rahul Pandit
		Link  : https://www.rahulpandit.com/post/file-encryption-using-xchaha20-poly1305-in-golang/

	At the time of this implementation, Rahul's article granted license for usage in this quote in the article:
		"I'm releasing this code into the public domain. You are free to use it however you want.
		If you find a bug, please let me know. My email address is listed on the about page. Thank you!"
*/

const (
	SaltLen    = 32
	KeyLen     = uint32(32)
	KeyTime    = uint32(5)
	KeyMemory  = uint32(64 * 1024)
	KeyThreads = uint8(4)
)

type ChachaCipher struct {
	ChunkSize    int
	DerivedKey   []byte
	Salt         []byte
	chacha       cipher.AEAD
	BytesWritten int
	BytesRead    int
}

func NewChaChaCipherRandomSalt(key []byte, chunkSize int) (*ChachaCipher, error) {
	salt := make([]byte, SaltLen)
	_, err := cryptorand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed creating random salt: %w", err)
	}

	return NewChaChaCipherFromSalt(key, salt, chunkSize)
}

func NewChaChaCipherFromSalt(key, salt []byte, chunkSize int) (*ChachaCipher, error) {
	c := &ChachaCipher{ChunkSize: chunkSize}

	c.deriveKey(key, salt)
	var err error
	c.chacha, err = chacha20poly1305.NewX(c.DerivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed creating chacha encrypter: %w", err)
	}
	return c, nil
}

func (c *ChachaCipher) GetBytesRead() int {
	return c.BytesRead
}
func (c *ChachaCipher) GetBytesWritten() int {
	return c.BytesWritten
}

func (c *ChachaCipher) GetChunkSize() int {
	return c.ChunkSize
}

func (c *ChachaCipher) GetDerivedKey() []byte {
	return c.DerivedKey
}

func (c *ChachaCipher) GetSalt() []byte {
	return c.Salt
}

func (c *ChachaCipher) deriveKey(keyIn, saltIn []byte) {
	c.Salt = make([]byte, len(saltIn))
	copy(c.Salt, saltIn)
	c.DerivedKey = argon2.IDKey(keyIn, c.Salt, KeyTime, KeyMemory, KeyThreads, KeyLen)
}

func (c *ChachaCipher) Decrypt(r io.Reader, w io.Writer) (int, error) {
	buffSize := c.chacha.NonceSize() + c.ChunkSize + c.chacha.Overhead()
	buf := make([]byte, buffSize)
	chunkCount := 1 // Used for error messages and as the AD value.  Starting at 1 is clearer in error messages.
	for {
		bytesRead, readErr := r.Read(buf)
		if bytesRead > 0 {
			c.BytesRead += bytesRead
			chunkBytes := buf[:bytesRead]
			if len(chunkBytes) < c.chacha.NonceSize() {
				return c.BytesWritten, fmt.Errorf(
					"failed reading input chunk %d: input size of %d is smaller than nonce size of %d",
					chunkCount,
					len(chunkBytes),
					c.chacha.NonceSize(),
				)
			}

			nonce, msgBytesEncrypted := chunkBytes[:c.chacha.NonceSize()], chunkBytes[c.chacha.NonceSize():]

			// Decrypt and validate
			msgBytesDecrypted, err := c.chacha.Open(nil, nonce, msgBytesEncrypted, []byte(strconv.Itoa(chunkCount)))
			if err != nil {
				return c.BytesWritten, fmt.Errorf("decrypt failed for stream in chunk %d: %w", chunkCount, err)
			}

			outputBytesWritten, outputErr := w.Write(msgBytesDecrypted)
			if outputErr != nil {
				return c.BytesWritten, fmt.Errorf("error writing chunk %d to output: %s", chunkCount, outputErr)
			}

			c.BytesWritten += outputBytesWritten

			if outputBytesWritten != len(msgBytesDecrypted) {
				return c.BytesWritten, fmt.Errorf(
					"error writing chunk %d. Bytes written: %d. Expected: %d",
					chunkCount,
					outputBytesWritten,
					len(msgBytesDecrypted),
				)
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return c.BytesWritten, fmt.Errorf("error reading chunk %d: %w", chunkCount, readErr)
		}

		chunkCount += 1
	}

	return c.BytesWritten, nil
}

func (c *ChachaCipher) Encrypt(r io.Reader, w io.Writer) (int, error) {
	buf := make([]byte, c.ChunkSize)
	chunkCount := 1 // Used for error messages and as the AD value.  Starting at 1 is clearer in error messages.

	for {
		bytesRead, readErr := r.Read(buf)
		if bytesRead > 0 {
			c.BytesRead += bytesRead
			// Select a random nonce, and leave capacity for the ciphertext.
			nonce := make([]byte, c.chacha.NonceSize(), c.chacha.NonceSize()+bytesRead+c.chacha.Overhead())
			_, err := cryptorand.Read(nonce)
			if err != nil {
				return c.BytesWritten, fmt.Errorf("error while processing chunk %d: %w", chunkCount, err)
			}

			msgBytesInput := buf[:bytesRead]

			// Encrypt message and append the ciphertext to the nonce.
			msgBytesEncypted := c.chacha.Seal(nonce, nonce, msgBytesInput, []byte(strconv.Itoa(chunkCount)))
			outputBytesWritten, outputErr := w.Write(msgBytesEncypted)
			if outputErr != nil {
				return c.BytesWritten, fmt.Errorf("error writing chunk %d to output: %s", chunkCount, outputErr)
			}
			c.BytesWritten += outputBytesWritten

			if outputBytesWritten != len(msgBytesEncypted) {
				return c.BytesWritten, fmt.Errorf(
					"error writing chunk %d. Bytes written: %d. Expected: %d",
					chunkCount,
					outputBytesWritten,
					len(msgBytesEncypted),
				)
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return c.BytesWritten, fmt.Errorf("error reading chunk %d: %s", chunkCount, readErr)
		}

		chunkCount += 1
	}

	return c.BytesWritten, nil
}
