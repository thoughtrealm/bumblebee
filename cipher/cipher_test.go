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
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSymmetricCipher(t *testing.T) {
	cipher, err := NewSymmetricCipher([]byte("key"), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, cipher)

	assert.NotEmpty(t, cipher.GetSalt())
}

func TestNewSymmetricCipherFromSalt(t *testing.T) {
	cipher, err := NewSymmetricCipherFromSalt([]byte("key"), []byte("salt"), 32000)
	assert.Nil(t, err)
	assert.NotNil(t, cipher)
	assert.Equal(t, []byte("salt"), cipher.GetSalt())
}

func TestNewKPCipherDecrypter(t *testing.T) {
	kp, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, kp)

	seed, err := kp.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, seed)

	pubkey, err := kp.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, pubkey)

	cipher, err := NewKPCipherDecrypter(seed, pubkey)
	assert.Nil(t, err)
	assert.NotNil(t, cipher)
}

func TestNewKPCipherEncrypter(t *testing.T) {
	kp, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, kp)

	seed, err := kp.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, seed)

	pubkey, err := kp.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, pubkey)

	cipher, err := NewKPCipherEncrypter(pubkey, seed)
	assert.Nil(t, err)
	assert.NotNil(t, cipher)
}
