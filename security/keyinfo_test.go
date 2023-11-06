package security

import (
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"testing"
)

const SIGN_INPUT_TEXT = "My name is Werner Brandon.  My voice is my passport.  Verify me."

var SignInputBytes = []byte(SIGN_INPUT_TEXT)

func TestKeyInfo_Sign(t *testing.T) {
	kpSender, err := nkeys.CreateCurveKeys()
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, kpSender)

	seedSender, err := kpSender.Seed()
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, seedSender)

	kiSender, err := NewKeyInfo(false, KeyTypeSeed, "sender", seedSender)
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, kiSender)

	kpReceiver, err := nkeys.CreateCurveKeys()
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, kpReceiver)

	publicKeyReceiver, err := kpReceiver.PublicKey()
	if !assert.Nil(t, err) {
		return
	}
	assert.NotEmpty(t, publicKeyReceiver)

	kiReceiver, err := NewKeyInfo(false, KeyTypePublic, "receiver", []byte(publicKeyReceiver))

	randomSignatureBytes, err := kiSender.SignRandom(SignInputBytes)
	if !assert.Nil(t, err) {
		return
	}
	assert.NotNil(t, randomSignatureBytes)

	isOk, err := kiReceiver.VerifyRandomSignature(randomSignatureBytes)
	assert.Nil(t, err)
	assert.True(t, isOk)
}
