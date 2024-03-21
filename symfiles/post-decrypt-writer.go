package symfiles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

type HeaderLoadProcessorFunc func(header *SymFileHeader) (targetWriter io.Writer, err error)

type StreamDecryptProcessorState int

const (
	ProcessorStateInitialized StreamDecryptProcessorState = iota
	ProcessorStateSizeWasRead
	ProcessStateHeaderWasRead
)

const SizeUInt16 = 2

type postStreamDecryptProcessor struct {
	postStreamDecryptBuff   []byte
	targetWriter            io.Writer
	processorState          StreamDecryptProcessorState
	headerSize              int
	headerLoadProcessorFunc HeaderLoadProcessorFunc
	symFileHeader           *SymFileHeader
}

func newPostStreamDecryptProcessor(targetWriter io.Writer, headerLoadProcessorFunc HeaderLoadProcessorFunc) *postStreamDecryptProcessor {
	return &postStreamDecryptProcessor{
		headerLoadProcessorFunc: headerLoadProcessorFunc,
		targetWriter:            targetWriter,

		// While initialized is the zero value, we set it explicitly in case the constants or type are ever changed
		processorState: ProcessorStateInitialized,
	}
}

// Write traps and parses the decrypted output stream for the purpose of handling the sym file header.
// If stream is still in a state of handling the header preamble, and no error occurs, then we must
// always return len(p) as the byte count written out.  Otherwise, the caller can get confused by
// what appears to be a partial write to the output.  This can be detected by some writers as an error.
func (psdp *postStreamDecryptProcessor) Write(p []byte) (n int, err error) {
	if psdp.processorState == ProcessStateHeaderWasRead {
		return psdp.targetWriter.Write(p)
	}

	var bytesIn []byte
	if len(psdp.postStreamDecryptBuff) > 0 {
		bytesIn = append(psdp.postStreamDecryptBuff, bytes.Clone(p)...)
		psdp.postStreamDecryptBuff = nil
	} else {
		bytesIn = bytes.Clone(p)
	}

	// It is HIGHLY unlikely that we would ever receive a Write of just 1 or 2 bytes in this
	// writer trap.  But we'll provide for those scenarios, just in case it does occur.
	if psdp.processorState == ProcessorStateInitialized {
		if len(bytesIn) < SizeUInt16 {
			// this probably can never occur, but we handle it just in case
			psdp.postStreamDecryptBuff = bytes.Clone(bytesIn)
			return len(p), nil
		}

		psdp.headerSize = int(binary.BigEndian.Uint16(bytesIn[:SizeUInt16]))
		psdp.processorState = ProcessorStateSizeWasRead

		if len(bytesIn) == SizeUInt16 {
			return len(p), nil
		}

		// remove size bytes from start of stream
		bytesIn = bytesIn[SizeUInt16:]
	}

	if psdp.processorState == ProcessorStateSizeWasRead {
		if len(bytesIn) < psdp.headerSize {
			psdp.postStreamDecryptBuff = bytes.Clone(bytesIn)
			return len(p), nil
		}

		psdp.symFileHeader = &SymFileHeader{}
		err = msgpack.Unmarshal(bytesIn[:psdp.headerSize], psdp.symFileHeader)
		if err != nil {
			return 0, fmt.Errorf("error reading header data from input stream: %w", err)
		}

		if psdp.targetWriter == nil {
			psdp.targetWriter, err = psdp.headerLoadProcessorFunc(psdp.symFileHeader)
			if err != nil {
				return 0, fmt.Errorf("failed processing header data from input stream: %w", err)
			}
		}

		psdp.processorState = ProcessStateHeaderWasRead
		if len(bytesIn) == psdp.headerSize {
			return len(p), nil
		}

		bytesToSendToWriter := bytes.Clone(bytesIn[psdp.headerSize:])
		bytesSentToWriter, err := psdp.targetWriter.Write(bytesToSendToWriter)

		// Todo: This check is tricky.  There are scenarios where the output byte count may not
		// match the expected output, but no error has occurred or some unexpected error occurs.
		// Practically, those scenarios should never occur.  But if they do, it should as expected,
		// but it will likely result in confusing output messages and/or unexpected results.
		// Might need to re-think the logic in this header handling, maybe track the output count
		// or errors a bit differently.
		if bytesSentToWriter != len(bytesToSendToWriter) {
			return bytesSentToWriter, err
		}

		return len(p), err
	}

	return 0, fmt.Errorf("unknown input stream processor state: %d", psdp.processorState)
}
