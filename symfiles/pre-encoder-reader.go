package symfiles

import (
	"bytes"
	"io"
)

type preStreamEncoderReader struct {
	preStreamEncoderBuff []byte
	sourceReader         io.Reader
}

func newPreStreamEncoderReader(preStreamBytes []byte, sourceReader io.Reader) *preStreamEncoderReader {
	return &preStreamEncoderReader{preStreamEncoderBuff: bytes.Clone(preStreamBytes), sourceReader: sourceReader}
}

func (psr *preStreamEncoderReader) Read(p []byte) (n int, err error) {
	var preStreamBytesCopied int
	var sourceBytesCopied int

	if len(psr.preStreamEncoderBuff) > 0 {
		preStreamBytesCopied = copy(p, bytes.Clone(psr.preStreamEncoderBuff))

		if preStreamBytesCopied < len(psr.preStreamEncoderBuff) {
			psr.preStreamEncoderBuff = bytes.Clone(psr.preStreamEncoderBuff[preStreamBytesCopied:])
			return
		}

		preStreamEncoderBuffLen := len(psr.preStreamEncoderBuff)
		psr.preStreamEncoderBuff = nil

		if len(p) == preStreamEncoderBuffLen {
			return
		}

		// The read requested more info than the length of the header, so fill the rest of the
		// request buffer with data from the sourceReader
		sourceBytesCopied, err = psr.sourceReader.Read(p[preStreamBytesCopied:])
		return sourceBytesCopied + preStreamBytesCopied, err
	}

	// The header has already been read and emitted to the reader, so just pass through reads from the
	// source reader from now on.
	sourceBytesCopied, err = psr.sourceReader.Read(p)
	return sourceBytesCopied + preStreamBytesCopied, err
}
