package symfiles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const DEFAULT_CHUNK_SIZE = 64000
const DEFAULT_SALT_SIZE = 32
const SymFileHeader_SIZE = 35

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
	HeaderLen   uint16
	PayloadType SymFilePayload
	Salt        []byte
}

func NewSymFileHeader(saltIn []byte, payloadType SymFilePayload) (*SymFileHeader, error) {
	if len(saltIn) != DEFAULT_SALT_SIZE {
		return nil, fmt.Errorf("NewSymFileHeader-> Invalid salt length: %d. Expected: %d bytes", len(saltIn), DEFAULT_SALT_SIZE)
	}

	return &SymFileHeader{
		HeaderLen:   SymFileHeader_SIZE,
		PayloadType: payloadType,
		Salt:        bytes.Clone(saltIn),
	}, nil
}

func LoadSymFileHeader(r io.Reader) (*SymFileHeader, error) {
	bytesInHeaderLen := make([]byte, 2) // length of uint64
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
		HeaderLen:   headerLen,
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
	var (
		outBytes          []byte
		localBytesWritten int
	)

	outBytes = binary.BigEndian.AppendUint16(outBytes, sfh.HeaderLen)
	outBytes = append(outBytes, byte(sfh.PayloadType))
	outBytes = append(outBytes, sfh.Salt...)

	localBytesWritten, err = w.Write(outBytes)
	return int64(localBytesWritten), err
}
