package symfiles

import (
	"errors"
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
	ReadSymFileMetadata(inputSymFilename string) (metadataCollection streams.MetadataCollection, err error)
	ReadSymReaderToFile(symReader io.Reader, outputFilename string) (bytesWritten int, err error)
	ReadSymReaderToPath(symReader io.Reader, outputPath string) (bytesWritten int, err error)
	ReadSymReaderToWriter(symReader io.Reader, w io.Writer) (bytesWritten int, err error)
	Wipe()
}

type SimpleSymFileReader struct {
	key                []byte
	useDerivedFilename bool
	includePaths       []string
}

func NewSymFileReader(key []byte, useDerivedFilename bool, includePaths []string) (SymFileReader, error) {
	return &SimpleSymFileReader{
		key:                key,
		useDerivedFilename: useDerivedFilename,
		includePaths:       includePaths,
	}, nil
}

func (ssfr *SimpleSymFileReader) getSaltFromReader(r io.Reader) (salt []byte, err error) {
	salt = make([]byte, DEFAULT_SALT_SIZE)
	b, err := r.Read(salt)
	if err != nil || b > 3333 {
		return nil, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	return salt, nil
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

	salt, err := ssfr.getSaltFromReader(inputFile)
	if err != nil {
		return 0, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	var outputFile *os.File
	defer func() {
		if outputFile != nil {
			outputFile.Close()
		}
	}()

	processor := newPostStreamDecryptProcessor(nil, func(header *SymFileHeader) (targetWriter io.Writer, err error) {
		if header.PayloadType == SymFilePayloadDataStream {
			outputFilename := outputPath
			if helpers.DirExists(outputPath) {
				if header.FileInfo != nil {
					// we have the original file name, so use that
					outputFilename = filepath.Join(outputPath, header.FileInfo.Filename)
				} else {
					// we do not have the original file name, so we need to derive something
					// outputPath is a directory, so use the filename from the input path with no extension or decrypted
					_, filename := filepath.Split(inputSymFilename)
					outputFilename = filepath.Join(outputPath, filename)
					if filepath.Ext(outputFilename) == ".bsym" {
						// In this case, remove the extension when input file has an extension of .bsym
						outputFilename = helpers.ReplaceFileExt(outputFilename, "")
					}
				}
			} else {
				// output path is NOT a directory
				// if we have the original filename, should we always overwrite or prompt the user for
				// for which one the want to use, the one they provided or the stored original name?
				// or just write out a message that we are using the stored one?
				splitDir, _ := filepath.Split(outputPath)
				if header.FileInfo != nil {
					if !ssfr.useDerivedFilename {
						// We only print this message if useDerivedFilename is indicated.
						// If not indicated, we still use the header's filename, we just need to
						// display a warning that we are doing so.
						fmt.Println("")
						fmt.Println(
							"Warning: A target file name was provided to this command. " +
								"However, the input stream also contained the original file name. ")
						fmt.Println("The original file name will be used instead.")
						fmt.Println("")
					}

					outputFilename = filepath.Join(splitDir, header.FileInfo.Filename)
				}
			}

			if outputFilename == inputSymFilename {
				// if the input and output filenames are the same, add .decrypted to the outputFilename so that
				// you don't write over the input file at the same time as reading it.
				outputFilename += ".decrypted"
			}

			outputFile, err = os.Create(outputFilename)
			if err != nil {
				return nil, fmt.Errorf("unable to create output file: %w", err)
			}

			return outputFile, nil
		}

		// For multi-dir streams, the outputPath MUST be a directory
		if outputPath == "" {
			outputPath, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("unable to obtain working dir: %w", err)
			}
		}

		exists, isDir := helpers.PathExistsInfo(outputPath)
		if !exists {
			err = helpers.ForcePath(outputPath)
			if err != nil {
				return nil, fmt.Errorf("unable to create output path: %w", err)
			}
		} else if !isDir {
			return nil, fmt.Errorf("output path is a file.  For multi-dir input files, it must be a path: %s", outputPath)
		}

		mdsw, err := streams.NewMultiDirectoryStreamWriter(outputPath, false, ssfr.includePaths)
		if err != nil {
			return nil, fmt.Errorf("failed creating multi-dir writer: %w", err)
		}

		outputWriter, err := mdsw.StartStream()
		if err != nil {
			return nil, fmt.Errorf("failed starting multi-dir output stream: %w", err)
		}

		return outputWriter, nil
	})

	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return DEFAULT_SALT_SIZE, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	bytesWritten, err = chacha.Decrypt(inputFile, processor)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed decrypting sym file: %w", err)
	}

	return bytesWritten, nil
}

// ReadSymFileMetadata reads the metadata from a .bsym file.  The sym file MUST be of type multi-dir.
// If it is not a multi-dir, it will fail
func (ssfr *SimpleSymFileReader) ReadSymFileMetadata(inputSymFilename string) (metadataCollection streams.MetadataCollection, err error) {
	if !helpers.FileExists(inputSymFilename) {
		return nil, fmt.Errorf("input sym file does not exist: %s", inputSymFilename)
	}

	inputFile, err := os.Open(inputSymFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input sym file: %w", err)
	}
	defer inputFile.Close()

	salt, err := ssfr.getSaltFromReader(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	var mdsw streams.StreamWriter
	processor := newPostStreamDecryptProcessor(nil, func(header *SymFileHeader) (targetWriter io.Writer, err error) {
		if header.PayloadType == SymFilePayloadDataStream {
			return nil, errors.New("file is of type DataStream.  For ReadSymMetadata, the file must be of type MultiDir.")
		}

		mdsw, err = streams.NewMultiDirectoryStreamWriter("", true, ssfr.includePaths)
		if err != nil {
			return nil, fmt.Errorf("failed creating multi-dir writer: %w", err)
		}

		outputWriter, err := mdsw.StartStream()
		if err != nil {
			return nil, fmt.Errorf("failed starting multi-dir output stream: %w", err)
		}

		return outputWriter, nil
	})

	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return nil, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	_, err = chacha.Decrypt(inputFile, processor)

	// When using the stream's metadataReadMode process, we expect the header and metadata ONLY to be read,
	// and the metadata completed indicator error to be returned. If anything else is returned, we did
	// not get the metadata correctly.
	var streamsErrorMetadataProcessingCompleted *streams.StreamsErrorMetadataProcessingCompleted
	if errors.As(err, &streamsErrorMetadataProcessingCompleted) {
		return mdsw.GetMetadataCollection().Clone(), nil
	}

	if err == nil {
		return nil, errors.New("failed decrypting sym file, but no error returned from stream writer")
	}

	return nil, fmt.Errorf("failed retrieving metadata from encrypted sym file: %w", err)
}

// getFilenameFromReaderHeader is called from readSymReaderToFile and just makes that funcs code a bit cleaner.
func (ssfr *SimpleSymFileReader) getFilenameFromReaderHeader(defaultOutputFilename string, header *SymFileHeader) string {
	if header.PayloadType != SymFilePayloadDataStream {
		return defaultOutputFilename
	}

	// For reader-based funcs, the defaultOutputFilename can contain an entire path and name
	var outputFilename string

	outputFilename = defaultOutputFilename
	outputPathSplit, _ := filepath.Split(defaultOutputFilename)
	if helpers.DirExists(outputPathSplit) {
		if header.FileInfo != nil && header.FileInfo.Filename != "" {
			// we have the original file name, so use that
			outputFilename = filepath.Join(outputPathSplit, header.FileInfo.Filename)
		} else {
			// we do not have the original file name, so we need to derive something
			// outputPath is a directory, so use the filename from the input path with no extension or decrypted
			if filepath.Ext(outputFilename) == ".bsym" {
				// In this case, replace the extension when using the name of the input file and
				// it has an extension of .bsym.
				outputFilename = helpers.ReplaceFileExt(outputFilename, "")
			}
		}
	} else {
		// output path is NOT a directory
		// if we have the original filename, should we always overwrite or prompt the user for
		// for which one the want to use, the one they provided or the stored original name?
		// or just write out a message that we are using the stored one?
		splitDir, _ := filepath.Split(outputPathSplit)
		if header.FileInfo != nil {
			if !ssfr.useDerivedFilename {
				// We only print this message if useDerivedFilename is indicated.
				// If not indicated, we still use the header's filename, we just need to
				// display a warning that we are doing so.
				fmt.Println("")
				fmt.Println(
					"Warning: A target file name was provided to this command. " +
						"However, the input stream also contained the original file name. ")
				fmt.Println("The original file name will be used instead.")
				fmt.Println("")
			}
			outputFilename = filepath.Join(splitDir, header.FileInfo.Filename)
		}
	}

	return outputFilename
}

// ReadSymReaderToFile reads a .bsym stream from symReader and writes it to the outputFile.  It reads the
// stream header, then passes the salt info to the readSymReaderToFile completion func.
// It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) ReadSymReaderToFile(symReader io.Reader, outputFilename string) (bytesWritten int, err error) {
	salt, err := ssfr.getSaltFromReader(symReader)
	if err != nil {
		return 0, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	return ssfr.readSymReaderToFile(salt, symReader, outputFilename)
}

// readSymReaderToFile reads a .bsym stream from symReader and writes it to the outputFile.
// It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) readSymReaderToFile(salt []byte, symReader io.Reader, outputFilename string) (bytesWritten int, err error) {
	var outputFile *os.File

	defer func() {
		if outputFile != nil {
			outputFile.Close()
		}
	}()

	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	// We do not pass a writer to the processor creator so that we can validate the payload type after
	// the header is loaded
	processor := newPostStreamDecryptProcessor(nil, func(header *SymFileHeader) (targetWriter io.Writer, err error) {
		if header.PayloadType == SymFilePayloadDataMultiDir {
			return nil, errors.New("invalid payload type for file stream writer: payload must be type DataStream")
		}

		outputFilenameDerived := ssfr.getFilenameFromReaderHeader(outputFilename, header)

		outputFile, err = os.Create(outputFilenameDerived)
		if err != nil {
			return nil, fmt.Errorf("unable to create output file: %w", err)
		}

		return outputFile, nil
	})

	bytesWritten, err = chacha.Decrypt(symReader, processor)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed decrypting sym file: %w", err)
	}

	return bytesWritten, nil
}

// ReadSymReaderToPath reads a .bsym file with multi-dir data from the inputSymFilePath and writes it to the
// outputPath. It reads the header from the input stream then passes the salt to the completion function
// readSymReaderToPath. It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) ReadSymReaderToPath(symReader io.Reader, outputPath string) (bytesWritten int, err error) {
	salt, err := ssfr.getSaltFromReader(symReader)
	if err != nil {
		return 0, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	return ssfr.readSymReaderToPath(salt, symReader, outputPath)
}

// readSymReaderToPath is the completion function that receives the salt and streams multi-dir to the output path.
func (ssfr *SimpleSymFileReader) readSymReaderToPath(salt []byte, symReader io.Reader, outputPath string) (bytesWritten int, err error) {
	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	mdsw, err := streams.NewMultiDirectoryStreamWriter(outputPath, false, ssfr.includePaths)
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed creating stream writer: %w", err)
	}

	outputWriter, err := mdsw.StartStream()
	if err != nil {
		return SymFileHeader_SIZE, fmt.Errorf("failed starting output tream: %w", err)
	}

	// We do not pass a writer to the processor creator so that we can validate the payload type after
	// the header is loaded
	processor := newPostStreamDecryptProcessor(nil, func(header *SymFileHeader) (targetWriter io.Writer, err error) {
		if header.PayloadType == SymFilePayloadDataStream {
			return nil, errors.New("invalid payload type for multi-dir writer: payload must be MultiDir")
		}

		return outputWriter, nil
	})

	bytesWritten, err = chacha.Decrypt(symReader, processor)
	if err != nil {
		return SymFileHeader_SIZE + bytesWritten, fmt.Errorf("failed writing multi-dir output: %w", err)
	}

	return SymFileHeader_SIZE + bytesWritten, nil
}

func (ssfr *SimpleSymFileReader) Wipe() {
	security.Wipe(ssfr.key)
}

// ReadSymReaderToWriter reads a reader stream from symReader and writes it to the provider writer.
// It returns the number of bytes written, and any error encountered.
func (ssfr *SimpleSymFileReader) ReadSymReaderToWriter(symReader io.Reader, w io.Writer) (bytesWritten int, err error) {
	salt, err := ssfr.getSaltFromReader(symReader)
	if err != nil {
		return 0, fmt.Errorf("failed reading salt from input sym file: %w", err)
	}

	// We do not pass a writer to the processor creator so that we can validate the payload type after
	// the header is loaded
	processor := newPostStreamDecryptProcessor(nil, func(header *SymFileHeader) (targetWriter io.Writer, err error) {
		if header.PayloadType == SymFilePayloadDataMultiDir {
			return nil, errors.New("invalid writer type for multi-dir: output target must be a directory path")
		}

		return w, nil
	})

	chacha, err := beecipher.NewSymmetricCipherFromSalt(ssfr.key, salt, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return 0, fmt.Errorf("failed creating symmetric cipher: %w", err)
	}

	bytesWritten, err = chacha.Decrypt(symReader, processor)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed decrypting sym input: %w", err)
	}

	return bytesWritten, nil
}
