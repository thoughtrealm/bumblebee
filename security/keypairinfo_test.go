package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeyPairInfoFromSeeds(t *testing.T) {
	kpi := NewKeyPairInfoFromSeeds("testname", []byte("testcipherseed"), []byte("testsigningseed"))
	if !assert.NotNil(t, kpi) {
		return
	}

	assert.Equal(t, "testname", kpi.Name)
	assert.Equal(t, []byte("testcipherseed"), kpi.CipherSeed)
	assert.Equal(t, []byte("testsigningseed"), kpi.SigningSeed)
}

func TestKeyPairInfo_Clone(t *testing.T) {
	kpi := NewKeyPairInfoFromSeeds("testname", []byte("testcipherseed"), []byte("testsigningseed"))
	if !assert.NotNil(t, kpi) {
		return
	}

	kpiClone := kpi.Clone()
	if !assert.NotNil(t, kpiClone) {
		return
	}

	assert.Equal(t, kpi.Name, kpiClone.Name)
	assert.Equal(t, kpi.CipherSeed, kpiClone.CipherSeed)
	assert.Equal(t, kpi.SigningSeed, kpiClone.SigningSeed)
}
