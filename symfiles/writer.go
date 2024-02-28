package symfiles

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/streams"
	"io"
	"os"
	"path/filepath"
)

func (ssf *SimpleSymFile) WriteSymFileFromFile(key []byte, inputFilename, outputSymFileName string) (bytesWritten int, err error) {
	if len(key) == 0 {
		return 0, errors.New("no key provided.  A key is required.")
	}

	if !helpers.FileExists(inputFilename) {
		return 0, fmt.Errorf("input file does not exist: %s", inputFilename)
	}

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return 0, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	return ssf.WriteSymFileFromReader(key, inputFile, outputSymFileName, SymFilePayloadDataStream)
}

func (ssf *SimpleSymFile) WriteSymFileFromDirs(key []byte, inputDirs []string, outputSymFileName string) (bytesWritten int, err error) {
	if len(key) == 0 {
		return 0, errors.New("no key provided.  A key is required.")
	}

	streamReader, err := streams.NewMultiDirectoryStreamReader(streams.WithCompression())
	if err != nil {
		return 0, fmt.Errorf("failed creating streamReader: %w", err)
	}

	for _, dir := range inputDirs {
		isFound, isDir, err := helpers.FileExistsWithDetails(dir)
		if err != nil {
			return 0, fmt.Errorf("error occurred validating dir \"%s\": %w", dir, err)
		}

		if !isFound {
			return 0, fmt.Errorf("directory \"%s\" not found", dir)
		}

		if !isDir {
			return 0, fmt.Errorf("\"%s\" is a file, not a directory", dir)
		}

		_, err = streamReader.AddDir(dir, streams.WithEmptyPaths(), streams.WithItemDetails())
		if err != nil {
			return 0, fmt.Errorf("unable to add dir \"%s\" to symfile job: %w", dir, err)
		}
	}

	r, err := streamReader.StartStream()
	if err != nil {
		return 0, fmt.Errorf("failed initializing multi-dir stream: %w", err)
	}

	return ssf.WriteSymFileFromReader(key, r, outputSymFileName, SymFilePayloadDataMultiDir)
}

func (ssf *SimpleSymFile) WriteSymFileFromReader(key []byte, r io.Reader, outputSymFileName string, payloadType SymFilePayload) (bytesWritten int, err error) {
	useSynName := outputSymFileName
	ext := filepath.Ext(useSynName)
	if ext == "" {
		// If it has an extension, we leave it alone.  If it does not, we add one.
		useSynName += ".bsym"
	}

	outputSymFile, err := os.Create(useSynName)
	if err != nil {
		return 0, fmt.Errorf("failed creating output symfile: %w", err)
	}
	defer outputSymFile.Close()

	chacha, err := cipher.NewChaChaCipherRandomSalt(key, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return 0, fmt.Errorf("unable to initialize chacha cipher: %w", err)
	}

	newHeader, err := NewSymFileHeader(chacha.Salt, payloadType)
	if err != nil {
		return 0, fmt.Errorf("failed creating new header: %w", err)
	}

	headerBytesWritten, err := newHeader.WriteTo(outputSymFile)
	if err != nil {
		return 0, fmt.Errorf("failed writing header bytes: %w", err)
	}

	fileBytesWritten, err := chacha.Encrypt(r, outputSymFile)
	if err != nil {
		return 0, fmt.Errorf("error writing symfile data: %w", err)
	}

	return int(headerBytesWritten) + fileBytesWritten, nil
}
