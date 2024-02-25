package streams

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Todo: This read pattern is really inefficient.  The temp cache write ahead approach will work, but
// way over complicates the requisite functionality for stream writing. It requires too much state
// management to track the oputput states between requested Read() calls.
// A much better approach would be to make this a WriteTo() pattern, passing in some target writer,
// perhaps the cipher writer. In that pattern, the stream writing functionality could be reduced to
// a simple set of nested for loops that emit the output to the target writer.

type StreamReader interface {
	io.Reader
	AddDir(dir string, options ...DirectoryOption) (newTree Tree, err error)
	GetTotalBytesRead() int64
	StartStream() (io.Reader, error)
}

// StreamBlockType indicates the type or function of data being written to a section of the stream
type StreamBlockType uint8

// StreamBlockType constants DO NOT use iota, because these are written to a persistent file
// stream.  We do not want possible future changes to break our ability to read a previously
// written file stream.

const (
	StreamBlockTypeTreeData   StreamBlockType = 1
	StreamBlockTypeItemHeader StreamBlockType = 2
	StreamBlockTypeItemData   StreamBlockType = 3
)

// StreamBlockFlag indicates various properties of the block
type StreamBlockFlag uint8

// StreamBlockFlag constants DO NOT use iota, because these are written to a persistent file
// stream.  We do not want possible future changes to break our ability to read a previously
// written file stream.

const (
	StreamBlockFlagLastDataBlock StreamBlockFlag = 1
	StreamBlockFlagIsCompressed  StreamBlockFlag = 2
)

// General stream constants

const (
	StreamBlockDescriptorLength         = 7
	StreamBlockDescriptorCurrentVersion = 1
	StreamItemHeaderLength              = 9
	StreamItemHeaderCurrentVersion      = 1
	StreamBlockDefaultMaxLength         = 20000000
	StreamFileReadBlockSize             = 64000
)

type StreamBlockDescriptor struct {
	Version  uint8
	Flags    uint8
	DataType StreamBlockType
	Length   uint32
}

func NewStreamBlockDescriptorFromBytes(input []byte) *StreamBlockDescriptor {
	if len(input) != StreamBlockDescriptorLength {
		panic(fmt.Sprintf("invalid StreamBlockDescriptor length.  Expected %d bytes, got %d bytes",
			StreamBlockDescriptorLength,
			len(input)),
		)
	}

	return &StreamBlockDescriptor{
		Version:  input[0],
		Flags:    input[1],
		DataType: StreamBlockType(input[2]),
		Length:   binary.BigEndian.Uint32(input[3:]),
	}
}

func (sbd *StreamBlockDescriptor) ToBytes() []byte {
	descBytes := make([]byte, StreamBlockDescriptorLength)
	descBytes[0] = sbd.Version
	descBytes[1] = sbd.Flags
	descBytes[2] = uint8(sbd.DataType)
	binary.BigEndian.PutUint32(descBytes[3:], sbd.Length)
	return descBytes
}

func (sbd *StreamBlockDescriptor) HasFlag(f StreamBlockFlag) bool {
	if sbd.Flags&uint8(f) == uint8(f) {
		return true
	}

	return false
}

type ItemHeader struct {
	Version uint8
	ItemID  uint32
	DirID   uint32
}

func NewItemHeaderFromBytes(itemHeaderBytes []byte) *ItemHeader {
	if len(itemHeaderBytes) != StreamItemHeaderLength {
		panic(fmt.Sprintf("invalid ItemHeader length.  Expected %d bytes, got %d bytes",
			StreamItemHeaderLength,
			len(itemHeaderBytes)),
		)
	}

	return &ItemHeader{
		Version: itemHeaderBytes[0],
		ItemID:  binary.BigEndian.Uint32(itemHeaderBytes[1:]),
		DirID:   binary.BigEndian.Uint32(itemHeaderBytes[5:]),
	}
}

func (ih ItemHeader) ToBytes() []byte {
	ihBytes := make([]byte, StreamItemHeaderLength)
	ihBytes[0] = ih.Version
	binary.BigEndian.PutUint32(ihBytes[1:], ih.ItemID)
	binary.BigEndian.PutUint32(ihBytes[5:], ih.DirID)
	return ihBytes
}

type iteratorSendInfo struct {
	err  error
	data []byte
}

type emitterRequestInfo struct {
	bytesNeeded int
}

type emitterResponseInfo struct {
	err  error
	data []byte
}

type StreamOption func(streamReader *MultiDirectoryStreamReader)

func WithCompression() StreamOption {
	return func(streamReader *MultiDirectoryStreamReader) {
		streamReader.compressionEnabled = true
	}
}

type CollectorSyncChannel chan *iteratorSendInfo
type AbortStreamChannel chan any
type RequestForDataChannel chan *emitterRequestInfo
type ResponseDataChannel chan *emitterResponseInfo

type MultiDirectoryStreamReader struct {
	cachedBytes           []byte
	compressionEnabled    bool
	currentTreesListIndex int
	currentDirNode        int
	currentItemsNode      int
	Comp                  Compressor
	Trees                 []Tree
	file                  *os.File
	totalBytesRead        int64
	startTime             time.Time
	chanRequestForData    RequestForDataChannel
	chanResponseData      ResponseDataChannel
}

func NewMultiDirectoryStreamReader(options ...StreamOption) (StreamReader, error) {
	streamReader := &MultiDirectoryStreamReader{
		Trees: make([]Tree, 0),
	}

	for _, option := range options {
		option(streamReader)
	}

	if streamReader.compressionEnabled {
		comp, err := NewCompressor()
		if err != nil {
			return nil, err
		}

		streamReader.Comp = comp
	}

	return streamReader, nil
}

func NewMultiDirectoryStreamReaderFromDirs(dirs []string, dirOptions []DirectoryOption, streamOptions []StreamOption) (StreamReader, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories provided")
	}

	mdsr, err := NewMultiDirectoryStreamReader(streamOptions...)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		_, err = mdsr.AddDir(dir, dirOptions...)
		if err != nil {
			return nil, err
		}
	}

	return mdsr, nil
}

func (mdsr *MultiDirectoryStreamReader) GetTotalBytesRead() int64 {
	return mdsr.totalBytesRead
}

func (mdsr *MultiDirectoryStreamReader) AddDir(dir string, options ...DirectoryOption) (newTree Tree, err error) {
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

	// It is possible that this is called multiple times to effectively restart the stream after an error.
	// So, we explicitly set these values, even though they are the default values.
	mdsr.currentTreesListIndex = 0
	mdsr.currentDirNode = 0
	mdsr.currentItemsNode = 0
	mdsr.file = nil
	mdsr.totalBytesRead = 0
	mdsr.startTime = time.Now()
	mdsr.chanRequestForData = make(chan *emitterRequestInfo)
	mdsr.chanResponseData = make(chan *emitterResponseInfo)

	go mdsr.Controller() // start the iterator/collector/emitter machine

	return mdsr, nil
}

func (mdsr *MultiDirectoryStreamReader) Controller() {
	var chanCollectorSync = make(CollectorSyncChannel)
	var chanAbortStream = make(AbortStreamChannel)

	go mdsr.Collector(chanCollectorSync, chanAbortStream, mdsr.chanRequestForData, mdsr.chanResponseData)
	go mdsr.Iterator(chanCollectorSync, chanAbortStream)
}

func (mdsr *MultiDirectoryStreamReader) Iterator(chanCollectorSync CollectorSyncChannel, chanAbortStream AbortStreamChannel) {
	for treeIdx, tree := range mdsr.Trees {
		treeAsBytes, err := tree.ToBytes()
		if err != nil {
			chanCollectorSync <- &iteratorSendInfo{
				err: fmt.Errorf("failed serializing tree data for tree index %d: %w", treeIdx, err),
			}
			return
		}

		blockHeader := &StreamBlockDescriptor{
			Version:  StreamBlockDescriptorCurrentVersion,
			DataType: StreamBlockTypeTreeData,
			Length:   uint32(len(treeAsBytes)),
		}
		blockHeaderBytes := blockHeader.ToBytes()

		chanCollectorSync <- &iteratorSendInfo{
			data: append(blockHeaderBytes, treeAsBytes...),
		}

		// check for abort
		select {
		case <-chanAbortStream:
			return
		default:
			// ignore this
		}

		rootPath := tree.GetRootPath()
		var (
			itemNode *ItemNode
			file     *os.File
		)

		for itemIndex := 0; itemIndex < tree.ItemCount(); itemIndex++ {
			// encapsulate in block to allow for defer to close file when done with reading
			func() {
				itemNode, err = tree.GetItemNodeByIndex(itemIndex)
				if err != nil {
					chanCollectorSync <- &iteratorSendInfo{
						err: fmt.Errorf("failed retrieving tree node: %w", err),
					}
					return
				}

				dirNode := tree.GetDirNodeByID(itemNode.DirID)
				if dirNode == nil {
					chanCollectorSync <- &iteratorSendInfo{
						err: fmt.Errorf("failed retrieving dir node for DirID: %d", itemNode.DirID),
					}
					return
				}

				file, err = os.Open(filepath.Join(rootPath, dirNode.Path, itemNode.Name))
				if err != nil {
					chanCollectorSync <- &iteratorSendInfo{
						err: fmt.Errorf("failed opening itemNode related file: %w", err),
					}
					return
				}

				defer func() {
					// Todo: How do we want to handle an error here?  Log it?  In a defer like this,
					// we are not sure what state the process is in when this is called.  It might be
					// returning from a completed file send, or maybe still looping on items/trees.
					_ = file.Close()
				}()

				ih := ItemHeader{
					Version: StreamItemHeaderCurrentVersion,
					ItemID:  uint32(itemNode.ItemID),
					DirID:   uint32(itemNode.DirID),
				}
				ihBytes := ih.ToBytes()

				blockHeader = &StreamBlockDescriptor{
					Version:  StreamBlockDescriptorCurrentVersion,
					DataType: StreamBlockTypeItemHeader,
					Length:   uint32(len(ihBytes)),
				}
				blockHeaderBytes = blockHeader.ToBytes()

				chanCollectorSync <- &iteratorSendInfo{
					data: append(blockHeaderBytes, ihBytes...),
				}

				select {
				case <-chanAbortStream:
					return
				default:
					// ignore this
				}

				var (
					bytesIn        = make([]byte, StreamFileReadBlockSize)
					done           bool
					isCompressed   bool
					bytesOut       []byte
					errCompression error
					bytesRead      int
				)

				for done == false {
					bytesRead, err = file.Read(bytesIn)
					bytesOut = bytesIn[:bytesRead]

					if bytesRead != 0 {
						mdsr.totalBytesRead += int64(bytesRead)
						if mdsr.compressionEnabled {
							bytesOut, isCompressed, errCompression = mdsr.Comp.CompressData(bytesIn[:bytesRead])
							if errCompression != nil {
								chanCollectorSync <- &iteratorSendInfo{
									err: fmt.Errorf("failed compressing data block: %w", errCompression),
								}
								return
							}

							if isCompressed {
								bytesRead = len(bytesOut)
							}
						}
					}

					blockHeader = &StreamBlockDescriptor{
						Version:  StreamBlockDescriptorCurrentVersion,
						DataType: StreamBlockTypeItemData,
						Length:   uint32(bytesRead),
					}

					if isCompressed {
						blockHeader.Flags = uint8(StreamBlockFlagIsCompressed)
					}

					if err == io.EOF {
						blockHeader.Flags = blockHeader.Flags | uint8(StreamBlockFlagLastDataBlock)
					}

					blockHeaderBytes = blockHeader.ToBytes()

					// build byte stream with the block header followed by the block's data bytes
					sendInfo := &iteratorSendInfo{}
					sendInfo.data = append(blockHeaderBytes, bytesOut...)

					if err != nil && err != io.EOF {
						sendInfo.err = err
					}

					chanCollectorSync <- sendInfo
					if err != nil && err != io.EOF {
						// Due to error reading file, we will just abort, instead of waiting for
						// abort channel signal
						return
					}

					// check for abort
					select {
					case <-chanAbortStream:
						return
					default:
						// ignore this
					}

					if err == io.EOF {
						done = true
					}
				}
			}()
		}
	}

	// Send a 0-byte data block with an EOF error... collector or controller may close the abort channel
	chanCollectorSync <- &iteratorSendInfo{err: io.EOF}
}

func (mdsr *MultiDirectoryStreamReader) Collector(
	chanCollectorSync CollectorSyncChannel,
	chanAbortStream AbortStreamChannel,
	chanRequestForData RequestForDataChannel,
	chanResponseData ResponseDataChannel) {

	var data []byte
	var lastErr error

	for {
		select {
		case <-chanAbortStream:
			return

		case requestForData := <-chanRequestForData:
			if lastErr != nil {
				chanResponseData <- &emitterResponseInfo{err: lastErr}
				return
			}

			if requestForData.bytesNeeded > len(data) {
				// not enough data cached to service request so cycle on collector requests until
				// we have enough data
				for {
					sendInfo := <-chanCollectorSync

					if len(sendInfo.data) > 0 {
						data = append(data, sendInfo.data...)
					}

					if sendInfo.err != nil {
						lastErr = sendInfo.err
						sendLen := mdsr.getSmallest(len(data), requestForData.bytesNeeded)
						chanResponseData <- &emitterResponseInfo{err: lastErr, data: data[:sendLen]}
						return
					}

					if len(data) >= requestForData.bytesNeeded {
						break
					}
				}
			}

			// The following checks against bytesNeeded and len(data) could be combined in one <=,
			// but breaking apart checks for < and == allows for a bit of optimization
			if requestForData.bytesNeeded < len(data) {
				chanResponseData <- &emitterResponseInfo{
					err:  lastErr,
					data: data[:requestForData.bytesNeeded],
				}
				data = bytes.Clone(data[requestForData.bytesNeeded:])

				if lastErr != nil {
					return
				}
			}

			if requestForData.bytesNeeded == len(data) {
				chanResponseData <- &emitterResponseInfo{
					err:  lastErr,
					data: data,
				}
				data = nil

				if lastErr != nil {
					return
				}
			}
		}
	}
}

func (mdsr *MultiDirectoryStreamReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		// can this ever occur?
		return 0, nil
	}

	mdsr.chanRequestForData <- &emitterRequestInfo{len(p)}
	dataResponse := <-mdsr.chanResponseData
	var sendLen int
	if len(dataResponse.data) != 0 {
		sendLen = mdsr.getSmallest(len(p), len(dataResponse.data))
		copy(p, dataResponse.data[:sendLen])
	}

	return sendLen, dataResponse.err
}

func (mdsr *MultiDirectoryStreamReader) getSmallest(val1, val2 int) int {
	if val1 < val2 {
		return val1
	}

	// They are either equal or val2 is smallest.  Either way, val2 is correct.
	return val2
}
