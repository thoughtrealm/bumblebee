package symfiles

import (
	"errors"
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

type SymFileHeader struct {
	HeaderLen   uint16
	PayloadType SymFilePayload
	Salt        [DEFAULT_SALT_SIZE]byte
}

func NewSymFileHeader(salt [DEFAULT_SALT_SIZE]byte, payloadType SymFilePayload) *SymFileHeader {
	return &SymFileHeader{
		HeaderLen:   SymFileHeader_SIZE,
		PayloadType: payloadType,
		Salt:        salt,
	}
}

func LoadHeader(r io.Reader) *SymFileHeader {
	return nil
}

func (sfh *SymFileHeader) WriteTo(w io.Writer) (bytesWritten int64, err error) {
	return 0, errors.New("not implemented")
}
