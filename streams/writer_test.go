package streams

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"os"
	"testing"
)

func TestNewMultiDirectoryStreamWriter(t *testing.T) {
	testHelperMultiDirectoryStreamWriter(t, nil, nil)
}

func testHelperMultiDirectoryStreamWriter(t *testing.T, preProcessFilter PreProcessFilter, preWriteFilter PreWriteFilter) (err error, bytesRead, bytesWritten int) {
	// remove the current test path if it exists
	err = os.RemoveAll("testdir_out")
	if err != nil {
		t.Logf("Unable to remove current test output path testdir_out: %s", err)
		return err, 0, 0
	}

	mdsw, err := NewMultiDirectoryStreamWriter("testdir_out")
	assert.NotNil(t, mdsw)
	assert.Nil(t, err)

	if preProcessFilter != nil {
		mdsw.SetPreProcessFilter(preProcessFilter)
	}

	if preWriteFilter != nil {
		mdsw.SetPreWriteFilter(preWriteFilter)
	}

	w, err := mdsw.StartStream()
	assert.NotNil(t, w)
	assert.Nil(t, err)

	inputFile, err := os.Open("testdata")
	if err != nil {
		t.Logf("Failed opening input file: %s", err)
		return err, 0, 0
	}

	var (
		readBuff   = make([]byte, 64000)
		readErr    error
		writeErr   error
		nRead      int
		nWrite     int
		writeCount int
	)

	for {
		nRead, readErr = inputFile.Read(readBuff)
		if readErr != nil && readErr != io.EOF {
			t.Logf("error reading input file: %s", readErr)
			return readErr, mdsw.TotalBytesRead(), mdsw.TotalBytesWritten()
		}

		writeCount += 1
		nWrite, writeErr = w.Write(readBuff[:nRead])
		if writeErr != nil {
			t.Logf("error writing output file: %s", writeErr)
			return writeErr, mdsw.TotalBytesRead(), mdsw.TotalBytesWritten()
		}

		if nWrite < nRead {
			t.Logf("bytes written %d is less than bytes read %d", nWrite, nRead)
			return fmt.Errorf("bytes written %d is less than bytes read %d",
				nWrite, nRead), mdsw.TotalBytesRead(), mdsw.TotalBytesWritten()
		}

		if readErr == io.EOF {
			break
		}
	}

	err = inputFile.Close()
	assert.Nil(t, err)

	p := message.NewPrinter(language.English)

	t.Logf("Last write count   : %d", writeCount)

	t.Log(p.Sprintf("Total bytes read   : %d", mdsw.TotalBytesRead()))
	t.Log(p.Sprintf("Total bytes written: %d", mdsw.TotalBytesWritten()))

	return nil, mdsw.TotalBytesRead(), mdsw.TotalBytesWritten()
}
