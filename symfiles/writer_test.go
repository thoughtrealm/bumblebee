package symfiles

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

var writerTestKey = []byte("testkey")

func TestSimpleSymFile_WriteSymFileFromFileSmall(t *testing.T) {
	_, _ = testHelperWriteSymFileFromFileSmall(t)
}

func testHelperWriteSymFileFromFileSmall(t *testing.T) (bytesWritten int, err error) {
	symFileWriter, err := NewSymFileWriter(writerTestKey)
	assert.NotNil(t, symFileWriter)
	assert.Nil(t, err)

	inputFilePath := filepath.Join("..", "streams", "testdir", "Dir2", "test.rtf")
	bytesWritten, err = symFileWriter.WriteSymFileFromFile(inputFilePath, "output_file_small.bsym")
	assert.Nil(t, err)
	assert.NotZero(t, bytesWritten)

	t.Logf("Writer Bytes written: %d", bytesWritten)

	return bytesWritten, err
}

func TestSimpleSymFile_WriteSymFileFromFileLarge(t *testing.T) {
	_, _ = testHelperWriteSymFileFromFileLarge(t)
}

func testHelperWriteSymFileFromFileLarge(t *testing.T) (bytesWritten int, err error) {
	symFileWriter, err := NewSymFileWriter(writerTestKey)
	assert.NotNil(t, symFileWriter)
	assert.Nil(t, err)

	inputFilePath := filepath.Join("..", "streams", "testdir2", "test2dir1", "bumblebee_0.1.1_darwin_all.tar.gz")
	bytesWritten, err = symFileWriter.WriteSymFileFromFile(inputFilePath, "output_file_large.bsym")
	assert.Nil(t, err)
	assert.NotZero(t, bytesWritten)

	t.Logf("Writer Bytes written: %d", bytesWritten)

	return bytesWritten, err
}

func TestSimpleSymFile_WriteSymFileFromDirs(t *testing.T) {
	_, _ = testHelperWriteSymFileFromDirs(t)
}

func testHelperWriteSymFileFromDirs(t *testing.T) (bytesWritten int, err error) {
	symFileWriter, err := NewSymFileWriter(writerTestKey)
	assert.NotNil(t, symFileWriter)
	assert.Nil(t, err)

	inputDirPath1 := filepath.Join("..", "streams", "testdir")
	inputDirPath2 := filepath.Join("..", "streams", "testdir2")
	bytesWritten, err = symFileWriter.WriteSymFileFromDirs([]string{inputDirPath1, inputDirPath2}, "output_dirs.bsym")
	assert.Nil(t, err)
	assert.NotZero(t, bytesWritten)

	t.Logf("Bytes written: %d", bytesWritten)

	return bytesWritten, err
}
