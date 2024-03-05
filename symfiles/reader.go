package symfiles

import (
	"fmt"
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/thoughtrealm/bumblebee/streams"
	"io"
	"os"
	"path/filepath"
)

type SymFileReader interface {
	ReadSymFile(inputSymFilename, outputPath string) (bytesWritten int, err error)
	Wipe()
}

type SimpleSymFileReader struct {
	key []byte
}

func NewSymFileReader(key []byte) (SymFileReader, error) {
	return &SimpleSymFileReader{
		key: key,
	}, nil
}

// ReadSymFile reads a .bsym file.  If the sym file is of type file stream, then outputPath must be a file
// name.  If the sym file is of type multi-dir stream, then outputPath must be a path name.
func (ssfr *SimpleSymFileReader) ReadSymFile(inputSymFilename, outputPath string) (bytesWritten int, err error) {
	if !helpers.FileExists(inputSymFilename) {
		return 0, fmt.Errorf("input sym file does not exist: %s", inputSymFilename)
	}

	inputFile, err := os.Open(inputSymFilename)
	if err != nil {
		return 0, fmt.Errorf("unable to open input sym file: %w", err)
	}
	defer inputFile.Close()

	newHeader, err := LoadSymFileHeader(inputFile)
	if err != nil {
		return 0, err
	}

	if newHeader.PayloadType == SymFilePayloadDataStream {
		outputFilename := outputPath
		if helpers.DirExists(outputPath) {
			// outputPath is a directory, so use the filename from the input path with no extension or decrypted
			_, filename := filepath.Split(inputSymFilename)
			outputFilename = filepath.Join(outputPath, filename)
			if filepath.Ext(outputFilename) == ".bsym" {
				// In this case, remove extension when input file has an extension of .bsym
				outputFilename = helpers.ReplaceFileExt(outputFilename, "")
			}

			if outputFilename == inputSymFilename {
				// if the input and output filenames are the same, add .decrypted to the outputFilename so that
				// you don't write over the input file at the same time as reading it.
				outputFilename += ".decrypted"
			}
		}

		return ssfr.ReadSymReaderToFile(newHeader.Salt, inputFile, outputFilename)
	}

	// For multi-dir streams, the outputPath MUST be a directory
	exists, isDir := helpers.PathExistsInfo(outputPath)
	if !exists {
		err = helpers.ForcePath(outputPath)
		if err != nil {
			return SymFileHeader_SIZE, fmt.Errorf("unable to create output path: %w", err)
		}
	} else if !isDir {
		return SymFileHeader_SIZE, fmt.Errorf("output path is a file.  For multi-dir input files, it must be a path: %s", outputPath)
	}

	return ssfr.ReadSymReaderToPath(newHeader.Salt, inputFile, outputPath)
}

// ReadSymReaderToFile reads a .bsym file from the inputSymFilePath and writes it to the outputFile.
// It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) ReadSymReaderToFile(salt []byte, symReader io.Reader, outputFilename string) (bytesWritten int, err error) {
	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("unable to create output file: %w", err)
	}
	defer outputFile.Close()

	bytesRead, err := chacha.Decrypt(symReader, outputFile)
	if err != nil {
		return SymFileHeader_SIZE + bytesRead, fmt.Errorf("failed decrypting sym file: %w", err)
	}

	return SymFileHeader_SIZE + bytesRead, nil
}

// ReadSymReaderToPath reads a .bsym file with multi-dir data from the inputSymFilePath and writes it to the outputPath.
// It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) ReadSymReaderToPath(salt []byte, symReader io.Reader, outputPath string) (bytesWritten int, err error) {
	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	mdsw, err := streams.NewMultiDirectoryStreamWriter(outputPath)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating stream writer: %w", err)
	}

	outputWriter, err := mdsw.StartStream()
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed starting output tream: %w", err)
	}

	bytesWritten, err = chacha.Decrypt(symReader, outputWriter)
	if err != nil {
		return SymFileHeader_SIZE + bytesWritten, fmt.Errorf("failed writing multi-dir output: %w", err)
	}

	return SymFileHeader_SIZE + bytesWritten, nil
}

func (ssfr *SimpleSymFileReader) Wipe() {
	security.Wipe(ssfr.key)
}
