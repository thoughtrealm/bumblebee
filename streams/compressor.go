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

type BundleCompressor interface{}

type Compressor struct {
	compressor *zstd.Encoder
}

func NewCompressor() (BundleCompressor, error) {
	// var err error

	comp := &Compressor{}
	return comp, nil
	/*
		comp.compressor, err = zstd.NewWriter(nil)

		enc, err := zstd.NewWriter(out)
		if err != nil {
			return err
		}
		_, err = io.Copy(enc, in)
		if err != nil {
			enc.Close()
			return err
		}
		return enc.Close()
	*/
}
