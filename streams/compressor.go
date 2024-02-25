package streams

import (
	"github.com/klauspost/compress/zstd"
)

/*
	todo: do the following things for compression...
		- determine length of input stream... only compress if stream input size is < some size
		- for smaller compressions use the zstandard's EncodeAll call and indicate this in the stream header
		- move chunk size to either a stream header or the bundle header.
		- Up the bundle header and/or data version to indicate new functionality around stored streams,
		  compression header, chunk size indicator per stream, etc.  Continue to support v1 encodings?
		- Store the chunk size in the bundle header or a stream header so it persists with the stream for
		  future decodings.
		- use zstandard/zstd for compression
*/

type Compressor interface {
	CompressData(inputData []byte) (newData []byte, isCompressed bool, err error)
}

type BeeCompressor struct {
	compressor *zstd.Encoder
}

func NewCompressor() (Compressor, error) {
	var err error
	beecomp := &BeeCompressor{}
	beecomp.compressor, err = zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, err
	}

	return beecomp, nil
}

func (bc *BeeCompressor) CompressData(inputData []byte) (newData []byte, isCompressed bool, err error) {
	newData = bc.compressor.EncodeAll(inputData, nil)
	if len(newData) >= len(inputData) {
		return inputData, false, nil
	}

	return newData, true, nil
}
