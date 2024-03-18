package symfiles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

const DEFAULT_CHUNK_SIZE = 64000
const DEFAULT_SALT_SIZE = 32
const SymFileHeader_SIZE = 35
const HeaderVersion = 1

type SymFilePayload uint8

const (
	SymFilePayloadDataStream   SymFilePayload = 0
	SymFilePayloadDataMultiDir SymFilePayload = 1
)

func isValidPayloadType(byteVal uint8) bool {
	switch byteVal {
	case uint8(SymFilePayloadDataStream):
		return true
	case uint8(SymFilePayloadDataMultiDir):
		return true
	default:
		return false
	}
}

type SymFileHeader struct {
	Version     int16
	PayloadType SymFilePayload
	Salt        []byte `msgpack:"-"`
	FileInfo    *SourceFileInfo
}

func NewSymFileHeader(saltIn []byte, payloadType SymFilePayload, sourceFileInfo *SourceFileInfo) (*SymFileHeader, error) {
	if len(saltIn) != DEFAULT_SALT_SIZE {
		return nil, fmt.Errorf("NewSymFileHeader-> Invalid salt length: %d. Expected: %d bytes", len(saltIn), DEFAULT_SALT_SIZE)
	}

	return &SymFileHeader{
		Version:     HeaderVersion,
		PayloadType: payloadType,
		Salt:        bytes.Clone(saltIn),
		FileInfo:    sourceFileInfo,
	}, nil
}

func LoadSymFileHeader(r io.Reader) (*SymFileHeader, error) {
	bytesInHeaderLen := make([]byte, 2) // length of uint16
	_, err := io.ReadFull(r, bytesInHeaderLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read HeaderLen from stream: %w", err)
	}

	headerLen := binary.BigEndian.Uint16(bytesInHeaderLen)
	if headerLen != SymFileHeader_SIZE {
		return nil, fmt.Errorf("invalid header length.  Expected %d bytes, but header value indicated %d bytes",
			SymFileHeader_SIZE, headerLen)
	}

	bytesInPayloadType := make([]byte, 1) // length of uint64
	_, err = io.ReadFull(r, bytesInPayloadType)
	if err != nil {
		return nil, fmt.Errorf("failing reading header payload type:  %w", err)
	}

	if !isValidPayloadType(bytesInPayloadType[0]) {
		return nil, fmt.Errorf("invalid payload type.  Expected 0 or 1, but received %d", bytesInPayloadType[0])
	}

	header := &SymFileHeader{
		PayloadType: SymFilePayload(bytesInPayloadType[0]),
		Salt:        make([]byte, DEFAULT_SALT_SIZE),
	}

	_, err = io.ReadFull(r, header.Salt)
	if err != nil {
		// this error will catch when the salt read was not able to read at least the default salt size,
		// so no need to explicitly check the read length
		return nil, fmt.Errorf("failing reading salt sequence: %w", err)
	}

	return header, nil
}

func (sfh *SymFileHeader) WriteTo(w io.Writer) (bytesWritten int64, err error) {
	headerBytes, err := sfh.ToBytes()
	if err != nil {
		return 0, err
	}

	localBytesWritten, err := w.Write(headerBytes)
	return int64(localBytesWritten), err
}

func (sfh *SymFileHeader) ToBytes() ([]byte, error) {
	msgpackBytes, err := msgpack.Marshal(sfh)
	if err != nil {
		return nil, fmt.Errorf("error in SymFile header ToBytes(): %w", err)
	}

	var headerLen uint16 = uint16(len(msgpackBytes))

	outBytes := binary.BigEndian.AppendUint16(nil, headerLen)
	outBytes = append(outBytes, msgpackBytes...)
	return outBytes, nil
}
