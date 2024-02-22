package streams

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestRunAllTests(t *testing.T) {
	var tests = []struct {
		name          string
		dirs          []string
		dirOptions    []DirectoryOption
		streamOptions []StreamOption
	}{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

			totalBytesWritten := 0
			totalBytesRead := 0
			for {
				var bytesRead int
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

			t.Logf("Total bytes read   : %d", totalBytesRead)
			t.Logf("Total bytes written: %d", totalBytesWritten)

			assert.Equal(t, totalBytesRead, totalBytesWritten)
		})
	}
}
