package symfiles

import (
	"fmt"
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/streams"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type SymFileWriter interface {
	SetSourceFileInfoFromStat(fi fs.FileInfo)
	WriteSymFileFromFile(inputFilename, outputSymFileName string) (bytesWritten int, err error)
	WriteSymFileFromDirs(inputDirs []string, outputSymFileName string, metadata []*streams.MetadataItem) (bytesWritten int, err error)
	WriteSymFileToWriterFromReader(r io.Reader, w io.Writer, payloadType SymFilePayload) (bytesWritten int, err error)
}

type SourceFileInfo struct {
	Filename  string
	Filedate  string // in time.RFC3339 format
	Fileperms uint16
}

func NewSourceFileInfoFromStat(fi fs.FileInfo) *SourceFileInfo {
	return &SourceFileInfo{
		Filename:  fi.Name(),
		Filedate:  fi.ModTime().UTC().Format(time.RFC3339),
		Fileperms: uint16(uint32(fi.Mode()) & uint32(0x1FF)),
	}
}

type SimpleSymFileWriter struct {
	writeBuffer    []byte
	sourceFileInfo *SourceFileInfo
	sc             beecipher.Cipher
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

func (ssfw *SimpleSymFileWriter) SetSourceFileInfoFromStat(fi fs.FileInfo) {
	ssfw.sourceFileInfo = NewSourceFileInfoFromStat(fi)
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

func (ssfw *SimpleSymFileWriter) WriteSymFileFromDirs(inputDirs []string, outputSymFileName string, metadata []*streams.MetadataItem) (bytesWritten int, err error) {
	streamReader, err := streams.NewMultiDirectoryStreamReader(streams.WithCompression())
	if err != nil {
		return 0, fmt.Errorf("failed creating streamReader: %w", err)
	}

	if metadata != nil {
		mc := streamReader.GetMetadataCollection()
		for _, item := range metadata {
			err = mc.AddMetadataItem(item)
			if err != nil {
				return 0, fmt.Errorf("failed adding stream metadata: %w", err)
			}
		}
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

	newHeader, err := NewSymFileHeader(ssfw.sc.GetSalt(), payloadType, ssfw.sourceFileInfo)
	if err != nil {
		return 0, fmt.Errorf("failed creating new header: %w", err)
	}

	headerBytes, err := newHeader.ToBytes()
	if err != nil {
		return 0, fmt.Errorf("failed writing header bytes: %w", err)
	}

	psw := newPreStreamEncoderReader(headerBytes, r)

	// we first write the salt/IV directly to the output stream unencrypted/unencoded
	saltBytesWritten, err := outputSymFile.Write(ssfw.sc.GetSalt())
	if err != nil {
		return 0, fmt.Errorf("error writing salt/IV: %w", err)
	}

	// Now, encrypt the input stream to the output stream
	fileBytesWritten, err := ssfw.sc.Encrypt(psw, outputSymFile)
	if err != nil {
		return 0, fmt.Errorf("error writing symfile data: %w", err)
	}

	return saltBytesWritten + fileBytesWritten, nil
}

func (ssfw *SimpleSymFileWriter) WriteSymFileToWriterFromReader(r io.Reader, w io.Writer, payloadType SymFilePayload) (bytesWritten int, err error) {
	newHeader, err := NewSymFileHeader(ssfw.sc.GetSalt(), payloadType, ssfw.sourceFileInfo)
	if err != nil {
		return 0, fmt.Errorf("failed creating new header: %w", err)
	}

	headerBytes, err := newHeader.ToBytes()
	if err != nil {
		return 0, fmt.Errorf("failed getting header bytes: %w", err)
	}

	pser := newPreStreamEncoderReader(headerBytes, r)

	// we first write the salt/IV to the output stream
	saltBytesWritten, err := w.Write(ssfw.sc.GetSalt())
	if err != nil {
		return 0, fmt.Errorf("error writing salt/IV: %w", err)
	}

	fileBytesWritten, err := ssfw.sc.Encrypt(pser, w)
	if err != nil {
		return 0, fmt.Errorf("error writing symfile data: %w", err)
	}

	return saltBytesWritten + fileBytesWritten, nil
}
