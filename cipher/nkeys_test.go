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

func TestNKeysCipher_SuccessOnNormalOperation(t *testing.T) {
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

func TestNKeysCipher_FailOnDecryptWithAttackerSenderPublicKey(t *testing.T) {
	receiverKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, receiverKP)

	receiverSeed, err := receiverKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, receiverSeed)

	receiverPubKey, err := receiverKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, receiverPubKey)

	senderGoodKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, senderGoodKP)

	senderGoodSeed, err := senderGoodKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, senderGoodSeed)

	senderGoodPubKey, err := senderGoodKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, senderGoodPubKey)

	senderEvilKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, senderEvilKP)

	senderEvilSeed, err := senderEvilKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, senderEvilSeed)

	senderEvilPubKey, err := senderEvilKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, senderEvilPubKey)

	nkeysEncrypterEvil, err := NewNKeysCipherEncrypter(receiverPubKey, senderEvilSeed)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysEncrypterEvil)

	nkeysDecrypter, err := NewNKeysCipherDecrypter(receiverSeed, senderGoodPubKey)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysDecrypter)

	secretBytes := werner_bytes
	encryptReadBuffer := bytes.NewBuffer(secretBytes)
	encryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysEncrypterEvil.Encrypt(encryptReadBuffer, encryptWriteBuffer)
	assert.Nil(t, err)

	decryptReadBuffer := bytes.NewBuffer(encryptWriteBuffer.Bytes())
	decryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysDecrypter.Decrypt(decryptReadBuffer, decryptWriteBuffer)
	assert.NotNil(t, err)

	assert.NotEqual(t, secretBytes, decryptWriteBuffer.Bytes())
}

func TestNKeysCipher_FailOnDecryptWithIncorrectSenderPublicKey(t *testing.T) {
	receiverKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, receiverKP)

	receiverSeed, err := receiverKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, receiverSeed)

	receiverPubKey, err := receiverKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, receiverPubKey)

	senderGoodKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, senderGoodKP)

	senderGoodSeed, err := senderGoodKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, senderGoodSeed)

	senderGoodPubKey, err := senderGoodKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, senderGoodPubKey)

	senderEvilKP, err := nkeys.CreateCurveKeys()
	assert.Nil(t, err)
	assert.NotNil(t, senderEvilKP)

	senderEvilSeed, err := senderEvilKP.Seed()
	assert.Nil(t, err)
	assert.NotNil(t, senderEvilSeed)

	senderEvilPubKey, err := senderEvilKP.PublicKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, senderEvilPubKey)

	nkeysGoodEncrypter, err := NewNKeysCipherEncrypter(receiverPubKey, senderGoodSeed)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysGoodEncrypter)

	nkeysDecrypter, err := NewNKeysCipherDecrypter(receiverSeed, senderEvilPubKey)
	assert.Nil(t, err)
	assert.NotNil(t, nkeysDecrypter)

	secretBytes := werner_bytes
	encryptReadBuffer := bytes.NewBuffer(secretBytes)
	encryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysGoodEncrypter.Encrypt(encryptReadBuffer, encryptWriteBuffer)
	assert.Nil(t, err)

	decryptReadBuffer := bytes.NewBuffer(encryptWriteBuffer.Bytes())
	decryptWriteBuffer := bytes.NewBuffer(nil)
	_, err = nkeysDecrypter.Decrypt(decryptReadBuffer, decryptWriteBuffer)
	assert.NotNil(t, err)

	assert.NotEqual(t, secretBytes, decryptWriteBuffer.Bytes())
}

func TestSigningSuccess(t *testing.T) {
	signKeySender, err := nkeys.CreateUser()
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, signKeySender) {
		return
	}

	const signText = "My Name is Werner Branden"
	signature, err := signKeySender.Sign([]byte(signText))
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, signature) {
		return
	}

	publickey, err := signKeySender.PublicKey()
	if !assert.Nil(t, err) {
		return
	}

	verifyKeyReceiver, err := nkeys.FromPublicKey(publickey)
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, verifyKeyReceiver) {
		return
	}

	err = verifyKeyReceiver.Verify([]byte(signText), signature)
	assert.Nil(t, err)
}

func TestSigningFailWithWrongKey(t *testing.T) {
	signKeySenderGood, err := nkeys.CreateUser()
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, signKeySenderGood) {
		return
	}

	signKeySenderBad, err := nkeys.CreateUser()
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, signKeySenderBad) {
		return
	}

	const signText = "My Name is Werner Branden"
	signature, err := signKeySenderBad.Sign([]byte(signText))
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, signature) {
		return
	}

	publickey, err := signKeySenderGood.PublicKey()
	if !assert.Nil(t, err) {
		return
	}

	verifyKeyReceiver, err := nkeys.FromPublicKey(publickey)
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, verifyKeyReceiver) {
		return
	}

	err = verifyKeyReceiver.Verify([]byte(signText), signature)
	assert.NotNil(t, err)
}
