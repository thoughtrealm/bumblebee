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
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncryptAndDecryptBytes(t *testing.T) {
	testData := []byte("My voice is my passport. Verify me.")
	testKey := []byte("Test testKey")

	encryptedBytes, salt, err := EncryptBytes(testData, testKey)
	if !assert.Nil(t, err) {
		return
	}

	assert.NotNil(t, salt)
	assert.NotNil(t, encryptedBytes)

	decryptedBytes, err := DecryptBytes(encryptedBytes, testKey, salt)
	if !assert.Nil(t, err) {
		return
	}

	assert.NotNil(t, decryptedBytes)
	assert.Equal(t, testData, decryptedBytes)

	fmt.Printf("Len Test Data: %d\n", len(testData))
	fmt.Printf("Len Encrypted Data: %d\n", len(encryptedBytes))
	fmt.Printf("Encrypted Data: %x\n", encryptedBytes)
	fmt.Printf("Encrypted string: %q\n", encryptedBytes)
	fmt.Printf("Test Data: %s\n", string(testData))
	fmt.Printf("Decrypted Data: %s\n", string(decryptedBytes))
}
