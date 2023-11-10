package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const SIGN_INPUT_TEXT = "My name is Werner Brandon.  My voice is my passport.  Verify me."

var SignInputBytes = []byte(SIGN_INPUT_TEXT)

func TestKeyInfo_SignAndVerify(t *testing.T) {
	senderKPI, err := NewKeyPairInfoWithSeeds("sender")
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, senderKPI)

	cipherPublicKey, signingPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(t, err) {
		return
	}

	kiReceiver, err := NewKeyInfo("receiver", cipherPublicKey, signingPublicKey)

	randomSignatureBytes, err := senderKPI.SignRandom(SignInputBytes)
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, randomSignatureBytes)

	isOk, err := kiReceiver.VerifyRandomSignature(randomSignatureBytes)
	assert.Nil(t, err)
	assert.True(t, isOk)
}
