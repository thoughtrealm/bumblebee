package streams

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewMultiDirectoryStreamWriterInSingleWrite(t *testing.T) {
	mdsw, err := NewMultiDirectoryStreamWriter("testdir_out")
	assert.NotNil(t, mdsw)
	assert.Nil(t, err)

	w, err := mdsw.StartStream()
	assert.NotNil(t, w)
	assert.Nil(t, err)

	fileBytes, err := os.ReadFile("testdata")
	assert.NotNil(t, fileBytes)
	assert.Nil(t, err)

	n, err := w.Write(fileBytes)
	assert.Nil(t, err)

	assert.Equal(t, len(fileBytes), n)
	assert.Nil(t, err)
}
