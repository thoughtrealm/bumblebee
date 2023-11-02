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
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNKeysCipher(t *testing.T) {
	receiverKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, receiverKP)

	receiverSeed, err := receiverKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, receiverSeed)

	receiverPubKey, err := receiverKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, receiverPubKey)

	senderKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, senderKP)

	senderSeed, err := senderKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, receiverSeed)

	senderPubKey, err := senderKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, receiverPubKey)

	nkeysEncrypter, err := NewNKeysCipherEncrypter(receiverPubKey, senderSeed)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysEncrypter)

	nkeysDecrypter, err := NewNKeysCipherDecrypter(receiverSeed, senderPubKey)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysDecrypter)

	secretBytes := werner_bytes
	encryptReadBuffer := bytes.NewBuffer(secretBytes)
	encryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysEncrypter.Encrypt(encryptReadBuffer, encryptWriteBuffer)
	assert.Nil(t, err)

	decryptReadBuffer := bytes.NewBuffer(encryptWriteBuffer.Bytes())
	decryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysDecrypter.Decrypt(decryptReadBuffer, decryptWriteBuffer)
	assert.Nil(t, err)

	assert.Equal(t, secretBytes, decryptWriteBuffer.Bytes())
}
