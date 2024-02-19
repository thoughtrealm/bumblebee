package streams

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestNewMultiDirectoryStreamReader(t *testing.T) {
	mdsr, err := NewMultiDirectoryStreamReader()
	assert.NotNil(t, mdsr)
	assert.Nil(t, err)

	newTree, err := mdsr.AddDir("testdir", WithEmptyPaths())
	assert.NotNil(t, newTree)
	assert.Nil(t, err)

	r, err := mdsr.StartStream()
	assert.NotNil(t, r)
	assert.Nil(t, err)

	readBuff := make([]byte, 20000)
	totalBytesRead := 0
	bytesCache := bytes.NewBuffer(nil)
	for {
		var bytesRead int
		bytesRead, err = mdsr.Read(readBuff)
		if bytesRead > 0 {
			bytesCache.Write(readBuff[:bytesRead])
		}

		totalBytesRead += bytesRead

		if err == io.EOF {
			break
		}

		assert.Nil(t, err)
		if err != nil {
			break
		}
	}

	err = os.WriteFile("testdata", bytesCache.Bytes(), 0666)
	assert.Nil(t, err)

	t.Logf("Total bytes read: %d", totalBytesRead)
}
