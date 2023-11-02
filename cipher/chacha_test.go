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
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/thoughtrealm/bumblebee/helpers"
	"testing"
)

// TestChachaCipherSmallSecret tests an encrypt/decrypt cycle with a small input byte set
func TestChachaCipherSmallSecret(t *testing.T) {
	chachaEncrypter, err := NewChaChaCipherRandomSalt([]byte("verifyme"), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, chachaEncrypter)

	secretBytes := werner_bytes
	encryptReadBuffer := bytes.NewBuffer(secretBytes)
	encryptWriteBuffer := bytes.NewBuffer(nil)

	_, err = chachaEncrypter.Encrypt(encryptReadBuffer, encryptWriteBuffer)
	if !assert.Nil(t, err) {
		return
	}

	chachaDecrypter, err := NewChaChaCipherFromSalt([]byte("verifyme"), chachaEncrypter.GetSalt(), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, chachaDecrypter)

	decryptReadBuffer := bytes.NewBuffer(encryptWriteBuffer.Bytes())
	decryptWriteBuffer := bytes.NewBuffer(nil)

	_, err = chachaDecrypter.Decrypt(decryptReadBuffer, decryptWriteBuffer)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, secretBytes, decryptWriteBuffer.Bytes())
}

// TestChachaCipherSmallSecret tests an encrypt/decrypt cycle with a 10MB input byte set
func TestChachaCipherBigSecret(t *testing.T) {
	const inputSize = 10000000
	const blockSize = 1000

	chachaEncrypter, err := NewChaChaCipherRandomSalt([]byte("verifyme"), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, chachaEncrypter)

	secretBytes := make([]byte, 0, inputSize)
	blockCount := inputSize / blockSize

	// We read in random sets in smaller block sizes and aggregate into the secret input bytes for encrypting
	for i := 0; i < blockCount; i++ {
		blockBytes, err := helpers.GetRandomBytes(blockSize)
		if !assert.Nil(t, err) {
			return
		}

		secretBytes = append(secretBytes, blockBytes...)
	}

	encryptReader := bytes.NewReader(secretBytes)
	encryptWriteBuffer := bytes.NewBuffer(nil)

	_, err = chachaEncrypter.Encrypt(encryptReader, encryptWriteBuffer)
	if !assert.Nil(t, err) {
		return
	}

	chachaDecrypter, err := NewChaChaCipherFromSalt([]byte("verifyme"), chachaEncrypter.GetSalt(), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, chachaDecrypter)

	decryptReadBuffer := bytes.NewBuffer(encryptWriteBuffer.Bytes())
	decryptWriteBuffer := bytes.NewBuffer(nil)

	_, err = chachaDecrypter.Decrypt(decryptReadBuffer, decryptWriteBuffer)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, secretBytes, decryptWriteBuffer.Bytes())
}
