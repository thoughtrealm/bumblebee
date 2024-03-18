package symfiles

import (
	"io"
)

type HeaderLoadProcessor func(header SymFileHeader) (targetWriter io.Writer, err error)

type postStreamDecryptProcessor struct {
	postStreamDecryptBuff []byte
	targetWriter          io.Writer
	headerProcessed       bool
	headerLoadProcessor   HeaderLoadProcessor
}

func newPostStreamDecryptProcessor(headerLoadProcessor HeaderLoadProcessor) *postStreamDecryptProcessor {
	return &postStreamDecryptProcessor{headerLoadProcessor: headerLoadProcessor}
}

func (psdp *postStreamDecryptProcessor) Write(p []byte) (n int, err error) {
	if !psdp.headerProcessed {
		
	}

	return psdp.targetWriter.Write(p)
}
