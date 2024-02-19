package streams

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/helpers"
	"io"
	"os"
	"path/filepath"
)

const WriterReceiveBuffSize = 64000

type StreamWriter interface {
	io.Writer
	StartStream() (io.Writer, error)
	TotalBytesReceived() int
	TotalBytesWritten() int
}

type MultiDirectoryStreamWriter struct {
	rootPath           string
	decomp             Decompressor
	Trees              []Tree
	currentBlock       *StreamBlockDescriptor
	currentItemHeader  *ItemHeader
	currentFile        *os.File
	currentTree        Tree
	blockBytesReceived int
	totalBytesReceived int
	totalBytesWritten  int
	recvBuff           []byte
	totalFiles         int
}

func NewMultiDirectoryStreamWriter(rootPath string) (StreamWriter, error) {
	decomp, err := NewDecompressor()
	if err != nil {
		return nil, err
	}

	return &MultiDirectoryStreamWriter{
		rootPath: rootPath,
		decomp:   decomp,
		Trees:    make([]Tree, 0),
	}, nil
}

func (mdsw *MultiDirectoryStreamWriter) TotalBytesReceived() int {
	return mdsw.totalBytesReceived
}

func (mdsw *MultiDirectoryStreamWriter) TotalBytesWritten() int {
	return mdsw.totalBytesWritten
}

func (mdsw *MultiDirectoryStreamWriter) StartStream() (io.Writer, error) {
	// We explicitly write these 0 values in case StartStream is called after prior usage,
	// so that this is a reset on the stream writer.
	mdsw.blockBytesReceived = 0
	mdsw.totalBytesReceived = 0
	mdsw.totalBytesWritten = 0
	mdsw.totalFiles = 0
	mdsw.currentFile = nil
	mdsw.currentBlock = nil
	mdsw.currentItemHeader = nil
	mdsw.currentTree = nil
	mdsw.recvBuff = []byte{}

	return mdsw, nil
}

func (mdsw *MultiDirectoryStreamWriter) EndStream() error {
	// this should be called by consumers of this functionality when done reading bundle input
	return errors.New("EndStream not implemented")
}

func (mdsw *MultiDirectoryStreamWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		// should never happen, but just in case?
		return
	}

	mdsw.recvBuff = append(mdsw.recvBuff, p...)
	mdsw.totalBytesReceived += len(p)

	for {
		if mdsw.currentBlock == nil {
			// start of buff should be a StreamBlockDescriptor
			if len(mdsw.recvBuff) < StreamBlockDescriptorLength {
				return len(p), nil
			}

			mdsw.currentBlock = NewStreamBlockDescriptorFromBytes(mdsw.recvBuff[:StreamBlockDescriptorLength])
			mdsw.recvBuff = mdsw.recvBuff[StreamBlockDescriptorLength:]
			mdsw.blockBytesReceived = 0
		}

		if mdsw.currentBlock.DataType == StreamBlockTypeTreeData {
			if uint32(len(mdsw.recvBuff)) < mdsw.currentBlock.Len {
				// Not all tree data received yet, so return for now
				return len(p), nil
			}

			// We at least have enough data for the tree structure, so read it in now
			tree := NewDirectoryTree()
			err = tree.FromBytes(mdsw.recvBuff[:mdsw.currentBlock.Len])
			if err != nil {
				return len(p), fmt.Errorf("failed loading tree data: %w", err)
			}

			// processTreeData will iterate all dirs and create them in the target output folder
			mdsw.processTreeData(tree)

			blockLen := mdsw.currentBlock.Len

			mdsw.currentBlock = nil
			mdsw.currentTree = tree
			mdsw.Trees = append(mdsw.Trees, tree)

			if uint32(len(mdsw.recvBuff)) == blockLen {
				// The buffer was exactly the length of the tree data, so reset recvBuff, return
				// and wait for more data
				mdsw.recvBuff = []byte{}
				return len(p), nil
			}

			// There is more data in the recvBuff than the tree data, so remove it from
			// the recvBuff and continue to the next block check
			mdsw.recvBuff = mdsw.recvBuff[blockLen:]
			continue
		}

		if mdsw.currentBlock.DataType == StreamBlockTypeItemHeader {
			blockLen := mdsw.currentBlock.Len
			if uint32(len(mdsw.recvBuff)) < blockLen {
				// not enough data yet for the item header
				return len(p), nil
			}

			mdsw.currentItemHeader = NewItemHeaderFromBytes(mdsw.recvBuff[:blockLen])

			mdsw.currentBlock = nil

			// call processItemHeader to open the item's file and prepare for writing data
			err = mdsw.processItemHeader()
			if err != nil {
				return len(p), fmt.Errorf("failed processing item header: %w", err)
			}

			if uint32(len(mdsw.recvBuff)) == blockLen {
				// The recvBuff is exactly the length of the ItemHeader, so reset recvBuff,
				// return and wait for more data
				mdsw.recvBuff = []byte{}
				return len(p), nil
			}

			// There is more data in the recvBuff than just the item header, so remove it from the
			// recvBuff and continue to the next block check
			mdsw.recvBuff = mdsw.recvBuff[blockLen:]
			continue
		}

		if mdsw.currentBlock.DataType == StreamBlockTypeItemData {
			if uint32(len(mdsw.recvBuff)) < mdsw.currentBlock.Len {
				// not enough data yet for the item data block
				return len(p), nil
			}

			// Read the item data from the recvBuff and write it to the current file
			data := mdsw.recvBuff[:mdsw.currentBlock.Len]
			fileBytesWritten, err := mdsw.currentFile.Write(data)
			if err != nil {
				return len(p), fmt.Errorf("failed writing item data: %w", err)
			}

			mdsw.totalBytesWritten += fileBytesWritten

			// Check if all item data has been received for this file
			if mdsw.currentBlock.HasFlag(StreamBlockFlagLastDataBlock) {
				// Close the current file
				err = mdsw.currentFile.Close()
				if err != nil {
					return len(p), fmt.Errorf("failed closing item file: %w", err)
				}

				// Reset the current file and item header
				mdsw.currentFile = nil
				mdsw.currentItemHeader = nil
			}

			blockLen := mdsw.currentBlock.Len
			mdsw.currentBlock = nil

			if uint32(len(mdsw.recvBuff)) == blockLen {
				// The recvBuff is exactly the length of the ItemData, so reset recvBuff,
				// return and wait for more data
				mdsw.recvBuff = []byte{}
				return len(p), nil
			}

			// There is more data in the recvBuff than just the item data, so remove it from the
			// recvBuff and continue to the next block check
			mdsw.recvBuff = mdsw.recvBuff[blockLen:]
			continue
		}

		return 0,
			fmt.Errorf(
				"Unknown block type detected in stream data: %d",
				int(mdsw.currentBlock.DataType))
	}
}

func (mdsw *MultiDirectoryStreamWriter) processTreeData(tree Tree) error {
	dirNodes := tree.GetDirNodes()
	for _, dirNode := range dirNodes {
		targetPath := filepath.Join(mdsw.rootPath, dirNode.Path)
		err := helpers.ForcePath(targetPath)
		if err != nil {
			return fmt.Errorf(
				"Unable to create path in output folder. Path: \"%s\".  Error: %w",
				targetPath,
				err)
		}
	}

	return nil
}

func (mdsw *MultiDirectoryStreamWriter) processItemHeader() error {
	// Open the item's file for writing
	// Get the DirNode
	dirNode := mdsw.currentTree.GetDirNodeByID(int(mdsw.currentItemHeader.DirID))
	if dirNode == nil {
		return fmt.Errorf("could not locate directory node with DirID: %d, ItemID: %d",
			mdsw.currentItemHeader.DirID,
			mdsw.currentItemHeader.ItemID)
	}

	fileNode := mdsw.currentTree.GetItemNodeByID(int(mdsw.currentItemHeader.ItemID))
	if fileNode == nil {
		return fmt.Errorf("could not locate item node with ItemID: %d",
			mdsw.currentItemHeader.ItemID)
	}

	file, err := os.Create(filepath.Join(mdsw.rootPath, dirNode.Path, fileNode.Name))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	mdsw.currentFile = file
	return nil
}
