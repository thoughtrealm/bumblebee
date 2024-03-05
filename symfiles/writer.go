package symfiles

import (
	"fmt"
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/streams"
	"io"
	"os"
	"path/filepath"
)

type SymFileWriter interface {
	WriteSymFileFromFile(inputFilename, outputSymFileName string) (bytesWritten int, err error)
	WriteSymFileFromDirs(inputDirs []string, outputSymFileName string) (bytesWritten int, err error)
	WriteSymFileToWriterFromReader(r io.Reader, w io.Writer, payloadType SymFilePayload) (bytesWritten int, err error)
}

type SimpleSymFileWriter struct {
	sc beecipher.Cipher
}

func NewSymFileWriter(key []byte) (SymFileWriter, error) {
	newCipher, err := beecipher.NewSymmetricCipher(key, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return nil, err
	}

	return &SimpleSymFileWriter{
		sc: newCipher,
	}, nil
}

func (ssfw *SimpleSymFileWriter) WriteSymFileFromFile(inputFilename, outputSymFileName string) (bytesWritten int, err error) {
	if !helpers.FileExists(inputFilename) {
		return 0, fmt.Errorf("input file does not exist: %s", inputFilename)
	}

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return 0, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	return ssfw.WriteSymFileFromReader(inputFile, outputSymFileName, SymFilePayloadDataStream)
}

func (ssfw *SimpleSymFileWriter) WriteSymFileFromDirs(inputDirs []string, outputSymFileName string) (bytesWritten int, err error) {
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

	return ssfw.WriteSymFileFromReader(r, outputSymFileName, SymFilePayloadDataMultiDir)
}

func (ssfw *SimpleSymFileWriter) WriteSymFileFromReader(r io.Reader, outputSymFileName string, payloadType SymFilePayload) (bytesWritten int, err error) {
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

	newHeader, err := NewSymFileHeader(ssfw.sc.GetSalt(), payloadType)
	if err != nil {
		return 0, fmt.Errorf("failed creating new header: %w", err)
	}

	headerBytesWritten, err := newHeader.WriteTo(outputSymFile)
	if err != nil {
		return 0, fmt.Errorf("failed writing header bytes: %w", err)
	}

	fileBytesWritten, err := ssfw.sc.Encrypt(r, outputSymFile)
	if err != nil {
		return 0, fmt.Errorf("error writing symfile data: %w", err)
	}

	return int(headerBytesWritten) + fileBytesWritten, nil
}

func (ssfw *SimpleSymFileWriter) WriteSymFileToWriterFromReader(r io.Reader, w io.Writer, payloadType SymFilePayload) (bytesWritten int, err error) {
	newHeader, err := NewSymFileHeader(ssfw.sc.GetSalt(), payloadType)
	if err != nil {
		return 0, fmt.Errorf("failed creating new header: %w", err)
	}

	headerBytesWritten, err := newHeader.WriteTo(w)
	if err != nil {
		return 0, fmt.Errorf("failed writing header bytes: %w", err)
	}

	fileBytesWritten, err := ssfw.sc.Encrypt(r, w)
	if err != nil {
		return 0, fmt.Errorf("error writing symfile data: %w", err)
	}

	return int(headerBytesWritten) + fileBytesWritten, nil
}
