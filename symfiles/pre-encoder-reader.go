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
	if len(psr.preStreamEncoderBuff) > 0 {
		bytesCopied := copy(p, psr.preStreamEncoderBuff)

		if bytesCopied < len(psr.preStreamEncoderBuff) {
			psr.preStreamEncoderBuff = psr.preStreamEncoderBuff[bytesCopied:]
			return
		}

		if len(p) == len(psr.preStreamEncoderBuff) {
			psr.preStreamEncoderBuff = nil
			return
		}

		// The read requested more info than the length of the header, so fill the rest of the
		// request buffer with data from the sourceReader
		return psr.sourceReader.Read(p[bytesCopied:])
	}

	// The header has already been read and emitted to the reader, so just pass through reads from the
	// source reader from now on.
	return psr.sourceReader.Read(p)
}
