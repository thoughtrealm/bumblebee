package streams

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestNewMultiDirectoryStreamWriter(t *testing.T) {
	// remove the current test path if it exists
	err := os.RemoveAll("testdir_out")
	if err != nil {
		t.Fatalf("Unable to remove current test output path testdir_out: %s", err)
	}

	mdsw, err := NewMultiDirectoryStreamWriter("testdir_out")
	assert.NotNil(t, mdsw)
	assert.Nil(t, err)

	w, err := mdsw.StartStream()
	assert.NotNil(t, w)
	assert.Nil(t, err)

	inputFile, err := os.Open("testdata")
	if err != nil {
		t.Fatalf("Failed opening input file: %s", err)
	}

	readBuff := make([]byte, 64000)
	for {
		n, err := inputFile.Read(readBuff)
		w.Write(readBuff[:n])

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("error reading input file: %s", err)
		}
	}

	err = inputFile.Close()
	assert.Nil(t, err)

	t.Logf("Total bytes read   : %d", mdsw.TotalBytesRead())
	t.Logf("Total bytes written: %d", mdsw.TotalBytesWritten())
}
