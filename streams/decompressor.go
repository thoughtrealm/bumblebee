package streams

import (
	"github.com/klauspost/compress/zstd"
)

type Decompressor interface {
	DecompressData(dataIn []byte) (dataOut []byte, err error)
}

type BeeDecompressor struct {
	decompressor *zstd.Decoder
}

func NewDecompressor() Decompressor {
	decomp := &BeeDecompressor{}
	decomp.decompressor, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	return decomp
}

func (bd *BeeDecompressor) DecompressData(dataIn []byte) (dataOut []byte, err error) {
	return bd.decompressor.DecodeAll(dataIn, nil)
}
