package streams

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thoughtrealm/bumblebee/helpers"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type testInfo struct {
	name             string
	dirs             []string
	dirOptions       []DirectoryOption
	streamOptions    []StreamOption
	preProcessFilter PreProcessFilter
	preWriteFilter   PreWriteFilter
	validateError    bool
	writeError       bool
}

func TestRunAllTests(t *testing.T) {
	var tests = []testInfo{
		{
			name:          "SingleDirWithEmptyPathsNoCompression",
			dirs:          []string{"testdir"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{},
		},
		{
			name:          "SingleDirWithEmptyPathsWithCompression",
			dirs:          []string{"testdir"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{WithCompression()},
		},
		{
			name:          "MultipleDirsWithEmptyPathsNoCompression",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{},
		},
		{
			name:          "MultipleDirsWithEmptyPathsWithCompression",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{WithCompression()},
		},
		{
			// This test SHOULD result in incorrect size comparisons during validation
			name:          "MultipleDirsWithEmptyPathsWithCompressionWithInvalidAdditionalDataWrittenOut",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{WithCompression()},
			preWriteFilter: func(dataIn []byte) (dataOut []byte) {
				newData := bytes.Clone(dataIn)
				return append(newData, []byte("create bad file lengths")...)
			},
			validateError: true,
			writeError:    false,
		},
		{
			// This test should result in decompression failures
			name:          "MultipleDirsWithEmptyPathsWithCompressionWithInvalidProcessingInput",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{WithEmptyPaths()},
			streamOptions: []StreamOption{WithCompression()},
			preProcessFilter: func(dataIn []byte) (dataOut []byte) {
				dataOut = bytes.Clone(dataIn)
				for idx := 0; idx < len(dataOut); idx++ {
					dataOut[idx] = dataIn[idx] ^ uint8(idx%256)
				}

				return dataOut
			},
			writeError: true,
		},
		{
			name:          "SingleDirNoEmptyPathsNoCompression",
			dirs:          []string{"testdir"},
			dirOptions:    []DirectoryOption{},
			streamOptions: []StreamOption{},
		},
		{
			name:          "SingleDirNoEmptyPathsWithCompression",
			dirs:          []string{"testdir"},
			dirOptions:    []DirectoryOption{},
			streamOptions: []StreamOption{WithCompression()},
		},
		{
			name:          "MultipleDirsNoEmptyPathsNoCompression",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{},
			streamOptions: []StreamOption{},
		},
		{
			name:          "MultipleDirsNoEmptyPathsWithCompression",
			dirs:          []string{"testdir", "testdir2"},
			dirOptions:    []DirectoryOption{},
			streamOptions: []StreamOption{WithCompression()},
		},
	}

	var readerBytesRead, readerBytesWritten int
	var writerBytesRead, writerBytesWritten int
	var allBytesRead, allBytesWritten int

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readerBytesRead, readerBytesWritten = testHelperMultiDirectoryStreamReader(test, t)

			var err error
			err, writerBytesRead, writerBytesWritten = testHelperMultiDirectoryStreamWriter(
				t, test.preProcessFilter, test.preWriteFilter)

			allBytesRead += readerBytesRead + writerBytesRead
			allBytesWritten += readerBytesWritten + writerBytesWritten

			if test.writeError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			if err != nil {
				// No need to check for correct output writes if writer returned an error
				return
			}

			outputCorrect := testHelperValidateOutput(test.dirs, test.dirOptions)
			if outputCorrect {
				t.Logf("validate: Dirs, item sizes and hashes matched")
			}

			if !test.validateError {
				assert.True(t, outputCorrect)
			} else {
				assert.False(t, outputCorrect)
			}
		})
	}

	p := message.NewPrinter(language.English)

	t.Log(p.Sprintf("All tests bytes read   : %d", allBytesRead))
	t.Log(p.Sprintf("All tests bytes written: %d", allBytesWritten))
}

func testHelperMultiDirectoryStreamReader(test testInfo, t *testing.T) (returnBytesRead, returnBytesWritten int) {
	mdsr, err := NewMultiDirectoryStreamReader(test.streamOptions...)
	assert.NotNil(t, mdsr)
	assert.Nil(t, err)

	for _, dir := range test.dirs {
		newTree, err := mdsr.AddDir(dir, test.dirOptions...)
		assert.NotNil(t, newTree)
		assert.Nil(t, err)
	}

	r, err := mdsr.StartStream()
	assert.NotNil(t, r)
	assert.Nil(t, err)

	readBuff := make([]byte, 64000)
	outputFile, err := os.Create("testdata")
	if err != nil {
		t.Fatalf("Failed opening output file: %s", err)
	}

	var (
		totalBytesWritten int
		totalBytesRead    int
		bytesRead         int
	)

	for {
		bytesRead, err = mdsr.Read(readBuff)
		if bytesRead > 0 {
			totalBytesRead += bytesRead
			bytesWritten, err := outputFile.Write(readBuff[:bytesRead])
			totalBytesWritten += bytesWritten

			if err != nil {
				t.Fatalf("Failed writing to output file: %s", err)
			}
		}

		if err == io.EOF {
			break
		}

		assert.Nil(t, err)
		if err != nil {
			break
		}
	}

	err = outputFile.Close()
	assert.Nil(t, err)

	p := message.NewPrinter(language.English)

	t.Log(p.Sprintf("Total bytes read   : %d", mdsr.GetTotalBytesRead()))
	t.Log(p.Sprintf("Total bytes written: %d", totalBytesWritten))

	assert.Equal(t, totalBytesRead, totalBytesWritten)

	return totalBytesRead, totalBytesWritten
}

func testHelperValidateOutput(sourceDirs []string, dirOptions []DirectoryOption) (success bool) {
	for _, sourceDir := range sourceDirs {
		if !iterateSourceDir(sourceDir, dirOptions) {
			return false
		}
	}

	return true
}

func iterateSourceDir(sourceDir string, dirOptions []DirectoryOption) (isOk bool) {
	dt, err := NewDirectoryTreeFromPath(sourceDir, dirOptions...)
	if err != nil {
		fmt.Printf("Error creating DirectoryTree: %s\n", err)
		return false
	}

	// Iterate all the paths first
	dirNodes := dt.GetDirNodes()
	for _, dirNode := range dirNodes {
		targetDir := filepath.Join("testdir_out", dt.GetParentPathPrefix(), dirNode.Path)
		if !helpers.DirExists(targetDir) {
			fmt.Printf("validate: path not found in target: %s\n", targetDir)
			return false
		}
	}

	// Next, iterate the files
	itemNodes := dt.GetItemNodes()
	for _, itemNode := range itemNodes {
		// need the dir node
		parentDirNode := dt.GetDirNodeByID(itemNode.DirID)
		if parentDirNode == nil {
			fmt.Printf("validate: parent dirNode not found in itemNode: %d, %s\n", itemNode.ItemID, itemNode.Name)
			return false
		}

		targetFile := filepath.Join("testdir_out", dt.GetParentPathPrefix(), parentDirNode.Path, itemNode.Name)
		if !helpers.FileExists(targetFile) {
			fmt.Printf("validate: target file not found: %s\n", targetFile)
			return false
		}

		sourceFile := filepath.Join(sourceDir, parentDirNode.Path, itemNode.Name)

		sourceStats, err := os.Stat(sourceFile)
		if err != nil {
			fmt.Printf("validate: failed reading source file stats: %s\n", err)
			return false
		}

		targetStats, err := os.Stat(targetFile)
		if err != nil {
			fmt.Printf("validate: failed reading target file stats: %s\n", err)
			return false
		}

		if sourceStats.Size() != targetStats.Size() {
			fmt.Printf("validate: target size does not match source size: %s\n", targetFile)
			return false
		}

		targetHash, err := helperGetFileHash(targetFile)
		if err != nil {
			fmt.Printf("validate: Failed hashing target file contents: %s, %s\n", targetFile, err)
			return false
		}

		sourceHash, err := helperGetFileHash(sourceFile)
		if err != nil {
			fmt.Printf("validate: Failed hashing source file contents: %s, %s\n", sourceFile, err)
			return false
		}

		if sourceHash != targetHash {
			fmt.Println("validate: source and target hash mismatch...")
			fmt.Printf("  Target file: %s\n", targetFile)
			fmt.Printf("  Source file: %s\n", sourceFile)
			fmt.Printf("  Target hash: %x\n", targetHash)
			fmt.Printf("  Source hash: %x\n", sourceHash)
			return false
		}
	}

	return true
}

func helperGetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed opening file: %s", err)
	}

	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", fmt.Errorf("failed generating hash from file stream: %s", err)
	}

	return string(hash.Sum(nil)), nil
}
