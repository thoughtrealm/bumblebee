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

import "io"

type Cipher interface {
	GetBytesRead() int
	GetBytesWritten() int
	GetChunkSize() int
	GetDerivedKey() []byte
	GetSalt() []byte
	Decrypt(r io.Reader, w io.Writer) (int, error)
	Encrypt(r io.Reader, w io.Writer) (int, error)
}

// NewSymmetricCipher would be called for encrypting when the salt needs to be derived.
func NewSymmetricCipher(key []byte, chunkSize int) (Cipher, error) {
	// Just using ChaCha/Poly with Argon2 for now.  The interface will allow us to
	// support different ciphers in the future if we wish to, as well as for creating mocks.
	return NewChaChaCipherRandomSalt(key, chunkSize)
}

// NewSymmetricCipherFromSalt would be called for decrypting when the salt was previously derived.
func NewSymmetricCipherFromSalt(key, salt []byte, chunkSize int) (Cipher, error) {
	// Just using ChaCha/Poly with Argon2 for now.  The interface will allow us to
	// support different ciphers in the future if we wish to, as well as for creating mocks.
	return NewChaChaCipherFromSalt(key, salt, chunkSize)
}

// NewKPCipherDecoder initializes an nkey encoding set
func NewKPCipherDecrypter(receiverSeed []byte, senderPubKey string) (Cipher, error) {
	// This interface invocation is really particular to the NKeys 25519 wrapper package.
	// We might refactor for something else in the future, so use the interface wrapper for now.
	// Will help with mocks as well.
	return NewNKeysCipherDecrypter(receiverSeed, senderPubKey)
}

// NewKPCipherEncoder initializes an nkey decoding set
func NewKPCipherEncrypter(receiverPubKey string, senderSeed []byte) (Cipher, error) {
	// This interface invocation is really particular to the NKeys 25519 wrapper package.
	// We might refactor for something else in the future, so use the interface wrapper for now.
	// Will help with mocks as well.
	return NewNKeysCipherEncrypter(receiverPubKey, senderSeed)
}
