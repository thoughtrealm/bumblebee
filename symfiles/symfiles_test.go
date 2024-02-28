package symfiles

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

var testKey = []byte("testkey")

func TestSimpleSymFile_WriteSymFileFromFile(t *testing.T) {
	symFile, err := NewSymFile(testKey)
	assert.NotNil(t, symFile)
	assert.Nil(t, err)

	inputFilePath := filepath.Join("..", "streams", "testdir", "Dir2", "test.rtf")
	bytesWritten, err := symFile.WriteSymFileFromFile(testKey, inputFilePath, "output.bsym")
	assert.Nil(t, err)
	assert.NotZero(t, bytesWritten)
}
