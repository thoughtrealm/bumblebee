package streams

import (
	"github.com/klauspost/compress/zstd"
)

type Decompressor interface{}

type BeeDecompressor struct {
	deompressor *zstd.Decoder
}

func NewDecompressor() (Decompressor, error) {
	// var err error

	decomp := &BeeDecompressor{}
	return decomp, nil
}
