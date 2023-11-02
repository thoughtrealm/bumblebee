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
	cryptorand "crypto/rand"
	"fmt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const BEE_AEAD_DATA = "*21!)4!"

// these are simple symmetric support funcs

func EncryptBytes(inputBytes, key []byte) (encryptedBytes, salt []byte, err error) {
	salt = make([]byte, SaltLen)
	_, err = cryptorand.Read(salt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating random salt: %w", err)
	}

	// Derive strong key using argon2
	derivedKey := argon2.IDKey(key, salt, KeyTime, KeyMemory, KeyThreads, KeyLen)

	// Initialize a chacha cipher
	chacha, err := chacha20poly1305.NewX(derivedKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating cipher: %w", err)
	}

	nonce := make([]byte, chacha.NonceSize(), chacha.NonceSize()+len(inputBytes)+chacha.Overhead())
	_, err = cryptorand.Read(nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("error while read nonce random bytes: %w", err)
	}

	encryptedBytes = chacha.Seal(nonce, nonce, inputBytes, []byte(BEE_AEAD_DATA))
	return encryptedBytes, salt, nil
}

func DecryptBytes(encryptedBytes, key, salt []byte) (decryptedBytes []byte, err error) {
	// Derive strong key using argon2
	derivedKey := argon2.IDKey(key, salt, KeyTime, KeyMemory, KeyThreads, KeyLen)

	// Initialize a chacha cipher
	chacha, err := chacha20poly1305.NewX(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	encryptedBuf := bytes.Clone(encryptedBytes)

	nonce, msgBytesEncrypted := encryptedBuf[:chacha.NonceSize()], encryptedBuf[chacha.NonceSize():]
	decryptedBytes, err = chacha.Open(nil, nonce, msgBytesEncrypted, []byte(BEE_AEAD_DATA))
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}

	return decryptedBytes, nil
}
