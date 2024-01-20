package streams

import (
	"fmt"
	"io"
)

type StreamReader interface {
	io.Reader
	AddDir(dir string, options ...DirectoryOption) (newTree Tree, err error)
	StartStream() (io.Reader, error)
}

const streamReaderBlockSize = 64000

type streamReaderState int

const (
	streamReaderStateNone streamReaderState = iota
	streamReaderStateReady
	streamReaderStateToBytes
	streamReaderStateItems
)

func streamReaderStateToText(state streamReaderState) string {
	switch state {
	case streamReaderStateNone:
		return "NONE"
	case streamReaderStateReady:
		return "READY"
	case streamReaderStateToBytes:
		return "TOBYTES"
	case streamReaderStateItems:
		return "ITEMS"
	default:
		return "ERROR: UNKNOWN STATE"
	}
}

type MultiDirectoryStreamReader struct {
	state                 streamReaderState
	tempBytes             []byte
	currentTreesListIndex int
	currentDirNode        int
	currentItemsNode      int
	outputBlock           []byte
	Comp                  Compressor
	Trees                 []Tree
}

func NewMultiDirectoryStreamReader() (StreamReader, error) {
	comp, err := NewCompressor()
	if err != nil {
		return nil, err
	}

	return &MultiDirectoryStreamReader{
		state: streamReaderStateNone, // set explicitly in case we change the constants in the future
		Comp:  comp,
		Trees: make([]Tree, 0),
	}, nil
}

func NewMultiDirectoryStreamReaderFromDirs(dirs []string, options ...DirectoryOption) (StreamReader, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories provided")
	}

	mdsr, err := NewMultiDirectoryStreamReader()
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		_, err = mdsr.AddDir(dir, options...)
		if err != nil {
			return nil, err
		}
	}

	return mdsr, nil
}

func (mdsr *MultiDirectoryStreamReader) AddDir(dir string, options ...DirectoryOption) (newTree Tree, err error) {
	if mdsr.state != streamReaderStateNone {
		return nil,
			fmt.Errorf(
				"invalid stream reader state for AddDir: %s.  Expected NONE",
				streamReaderStateToText(mdsr.state),
			)
	}

	newTree, err = NewDirectoryTreeFromPath(dir, options...)
	if err != nil {
		return nil, err
	}

	mdsr.Trees = append(mdsr.Trees, newTree)

	return newTree, nil
}

func (mdsr *MultiDirectoryStreamReader) StartStream() (io.Reader, error) {
	if len(mdsr.Trees) == 0 {
		return nil, fmt.Errorf("no directory trees loaded for reading")
	}

	mdsr.outputBlock = make([]byte, streamReaderBlockSize)
	mdsr.state = streamReaderStateReady

	// It is possible that this is called multiple times to effectively restart the stream after an error.
	// So, we explicitly set these values, even though they are the default values.
	mdsr.currentTreesListIndex = 0
	mdsr.currentDirNode = 0
	mdsr.currentItemsNode = 0

	return mdsr, nil
}

func (mdsr *MultiDirectoryStreamReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("requested 0 bytes from Read()")
	}

	// These calls will prep and populate the p argument
	switch mdsr.state {
	case streamReaderStateReady:
		return mdsr.prepOutputBlockFromReady(p)
	case streamReaderStateToBytes:
		return mdsr.prepOutputBlockFromToBytes(p)
	case streamReaderStateItems:
		return mdsr.prepOutputBlockFromItems(p)
	default:
		return 0, fmt.Errorf(
			"invalid state for Read(): %s. Expected READY, TOBYTES or ITEMS",
			streamReaderStateToText(mdsr.state),
		)
	}
}

func (mdsr *MultiDirectoryStreamReader) prepOutputBlockFromReady(p []byte) (int, error) {
	// READY means we haven't started yet.  So, this is the very beginning of reading
	// the tree list.

	// We have already validated that len(p) > 0.

	// we know there is at least one tree, due to prior checks and the current state
	// so no need to validate if there is at least one tree or not
	mdsr.currentTreesListIndex = 0
	mdsr.currentDirNode = 0
	mdsr.currentItemsNode = 0

	// The first step is to serialize the first tree we are going to read in
	mdsr.state = streamReaderStateToBytes
	toBytes, err := mdsr.Trees[0].ToBytes()
	if err != nil {
		return 0, err
	}

	if len(toBytes) > len(p) {
		// Copies up to the lenght of p
		copy(p, toBytes)
		mdsr.tempBytes = toBytes[len(p):]
		return len(p), nil
	}

	if len(toBytes) == len(p) {
		mdsr.state = streamReaderStateItems
		copy(p, toBytes)
		mdsr.currentDirNode = -1
		return len(p), err
	}

	// At this point, we know that toBytes is less than the size of the request block
	// So we need to fill it with the toBytes data, then fill the remaining buffer with
	// the first node stream.
	copy(p, toBytes)

	// Todo: Confirm that p[pos:] is actually pointing within that slice where I want
	mdsr.state = streamReaderStateItems
	bytesWritten, err := mdsr.prepOutputBlockFromItems(p[len(toBytes):])

	return len(toBytes) + bytesWritten, err
}

func (mdsr *MultiDirectoryStreamReader) prepOutputBlockFromToBytes(p []byte) (int, error) {
	return 0, nil
}

func (mdsr *MultiDirectoryStreamReader) prepOutputBlockFromItems(p []byte) (int, error) {
	return 0, nil
}
