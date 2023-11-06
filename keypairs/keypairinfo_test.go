package keypairs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeyPairInfo(t *testing.T) {
	kpi := NewKeyPairInfo("testname", []byte("testseed"))
	if !assert.NotNil(t, kpi) {
		return
	}

	assert.Equal(t, "testname", kpi.Name)
	assert.Equal(t, []byte("testseed"), kpi.Seed)
}

func TestKeyPairInfo_Clone(t *testing.T) {
	kpi := NewKeyPairInfo("testname", []byte("testseed"))
	if !assert.NotNil(t, kpi) {
		return
	}

	kpiClone := kpi.Clone()
	if !assert.NotNil(t, kpiClone) {
		return
	}

	assert.Equal(t, kpi.Name, kpiClone.Name)
	assert.Equal(t, kpi.Seed, kpiClone.Seed)
}
