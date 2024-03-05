package symfiles

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var readerTestKey = []byte("testkey")

func TestSimpleSymFile_ReadSymFile(t *testing.T) {
	type test struct {
		name       string
		inputPath  string
		outputPath string
	}

	tests := []test{
		{
			name:       "Small File",
			inputPath:  "output_file_small.bsym",
			outputPath: "testdir_out/test.rtf",
		},
		{
			name:       "Large File",
			inputPath:  "output_file_large.bsym",
			outputPath: "testdir_out/bumblebee_0.1.1_darwin_all.tar.gz",
		},
		{
			name:       "Dirs",
			inputPath:  "output_dirs.bsym",
			outputPath: "testdir_out",
		},
	}

	// remove the current test output path if it exists
	err := os.RemoveAll("testdir_out")
	if err != nil {
		t.Logf("Unable to remove current test output path testdir_out: %s", err)
		return
	}

	os.Remove("output_dirs.bsym")
	os.Remove("output_file_small.bsym")
	os.Remove("output_file_large.bsym")

	_ = os.Mkdir("testdir_out", os.ModePerm)

	_, err = testHelperWriteSymFileFromFileSmall(t)
	assert.Nil(t, err)

	if err != nil {
		t.FailNow()
	}

	_, err = testHelperWriteSymFileFromFileLarge(t)
	assert.Nil(t, err)

	if err != nil {
		t.FailNow()
	}

	_, err = testHelperWriteSymFileFromDirs(t)
	assert.Nil(t, err)

	if err != nil {
		t.FailNow()
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			symFile, err := NewSymFileReader(writerTestKey)
			assert.NotNil(t, symFile)
			assert.Nil(t, err)

			bytesWritten, err := symFile.ReadSymFile(tc.inputPath, tc.outputPath)
			assert.Nil(t, err)
			assert.NotZero(t, bytesWritten)

			t.Logf("Bytes written: %d", bytesWritten)
		})
	}
}
