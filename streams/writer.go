package streams

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/helpers"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const WriterReceiveBuffSize = 64000

// PreProcessFilter is a testing mechanism that allows for altering the data sequence before it is processed,
// such as decompressing
type PreProcessFilter func(dataIn []byte) (dataOut []byte)

// PreWriteFilter is a testing mechanism that allows for altering the output stream before it is written out
type PreWriteFilter func(dataIn []byte) (dataOut []byte)

type StreamsErrorMetadataProcessingCompleted struct{}

func (mcc *StreamsErrorMetadataProcessingCompleted) Error() string {
	return "metadata collection complete"
}

type StreamWriter interface {
	io.Writer
	GetMetadataCollection() MetadataCollection
	GetMetadataReadMode() bool
	SetPreProcessFilter(preProcessFilter PreProcessFilter)
	SetPreWriteFilter(preWriteFilter PreWriteFilter)
	TotalBytesRead() int
	TotalBytesWritten() int
	StartStream() (io.Writer, error)
}

type MultiDirectoryStreamWriter struct {
	rootPath                string
	decomp                  Decompressor
	Trees                   []Tree
	Metadata                MetadataCollection
	currentBlock            *StreamBlockDescriptor
	currentItemHeader       *ItemHeader
	currentFilePath         string
	currentFile             *os.File
	currentDirNode          *DirNode
	currentItemNode         *ItemNode
	currentTree             Tree
	blockBytesRead          int
	totalBytesRead          int
	totalBytesWritten       int
	recvBuff                []byte
	totalFiles              int
	preProcessFilter        PreProcessFilter
	preWriteFilter          PreWriteFilter
	requireConfirm          bool
	overwriteDenyAll        bool
	ignoreCurrentItemOutput bool
	metadataReadMode        bool
	metadataReadComplete    bool
	includePaths            []string
}

func NewMultiDirectoryStreamWriter(rootPath string, metadataReadMode bool, includePaths []string) (StreamWriter, error) {
	decomp := NewDecompressor()

	return &MultiDirectoryStreamWriter{
		rootPath:         rootPath,
		decomp:           decomp,
		Trees:            make([]Tree, 0),
		requireConfirm:   true,
		metadataReadMode: metadataReadMode,
		includePaths:     includePaths,
	}, nil
}

func (mdsw *MultiDirectoryStreamWriter) GetMetadataCollection() MetadataCollection {
	return mdsw.Metadata
}

func (mdsw *MultiDirectoryStreamWriter) GetMetadataReadMode() bool {
	return mdsw.metadataReadMode
}

func (mdsw *MultiDirectoryStreamWriter) SetPreWriteFilter(preWriteFilter PreWriteFilter) {
	mdsw.preWriteFilter = preWriteFilter
}

func (mdsw *MultiDirectoryStreamWriter) SetPreProcessFilter(preProcessFilter PreProcessFilter) {
	mdsw.preProcessFilter = preProcessFilter
}

func (mdsw *MultiDirectoryStreamWriter) TotalBytesRead() int {
	return mdsw.totalBytesRead
}

func (mdsw *MultiDirectoryStreamWriter) TotalBytesWritten() int {
	return mdsw.totalBytesWritten
}

func (mdsw *MultiDirectoryStreamWriter) StartStream() (io.Writer, error) {
	// We explicitly write these 0 values in case StartStream is called after prior usage,
	// so that this is a reset on the stream writer.
	mdsw.blockBytesRead = 0
	mdsw.totalBytesRead = 0
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

	if mdsw.metadataReadMode && mdsw.metadataReadComplete {
		return 0, &StreamsErrorMetadataProcessingCompleted{}
	}

	mdsw.recvBuff = append(mdsw.recvBuff, p...)
	mdsw.totalBytesRead += len(p)

	for {
		if mdsw.currentBlock == nil {
			// start of buff should be a StreamBlockDescriptor
			if len(mdsw.recvBuff) < StreamBlockDescriptorLength {
				return len(p), nil
			}

			mdsw.currentBlock = NewStreamBlockDescriptorFromBytes(mdsw.recvBuff[:StreamBlockDescriptorLength])
			mdsw.recvBuff = mdsw.recvBuff[StreamBlockDescriptorLength:]
			mdsw.blockBytesRead = 0
		}

		if mdsw.currentBlock.DataType == StreamBlockTypeMetadata {
			if uint32(len(mdsw.recvBuff)) < mdsw.currentBlock.Length {
				// Not all tree data received yet, so return for now
				return len(p), nil
			}

			mdsw.Metadata, err = NewMetaDataCollectionFromBytes(mdsw.recvBuff[:mdsw.currentBlock.Length])
			if err != nil {
				return len(p), fmt.Errorf("failed loading metadata: %w", err)
			}

			mdsw.metadataReadComplete = true

			if mdsw.metadataReadMode {
				return len(p), &StreamsErrorMetadataProcessingCompleted{}
			}

			blockLen := mdsw.currentBlock.Length

			mdsw.currentBlock = nil
			if uint32(len(mdsw.recvBuff)) == blockLen {
				// The buffer was exactly the length of the metadata, so reset recvBuff, return
				// and wait for more data
				mdsw.recvBuff = []byte{}
				return len(p), nil
			}

			// There is more data in the recvBuff than the metadata, so remove it from
			// the recvBuff and continue to the next block check
			mdsw.recvBuff = mdsw.recvBuff[blockLen:]
			continue
		}

		if mdsw.currentBlock.DataType == StreamBlockTypeTreeData {
			if uint32(len(mdsw.recvBuff)) < mdsw.currentBlock.Length {
				// Not all tree data received yet, so return for now
				return len(p), nil
			}

			// We at least have enough data for the tree structure, so read it in now
			tree := NewDirectoryTree()
			err = tree.FromBytes(mdsw.recvBuff[:mdsw.currentBlock.Length])
			if err != nil {
				return len(p), fmt.Errorf("failed loading tree data: %w", err)
			}

			// processTreeData will iterate all dirs and create them in the target output folder
			err = mdsw.processTreeData(tree)
			if err != nil {
				return len(p), fmt.Errorf("failed processing tree data: %w", err)
			}

			blockLen := mdsw.currentBlock.Length

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
			blockLen := mdsw.currentBlock.Length
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
			if uint32(len(mdsw.recvBuff)) < mdsw.currentBlock.Length {
				// not enough data yet for the item data block
				return len(p), nil
			}

			// Read the item data from the recvBuff and write it to the current file
			data := mdsw.recvBuff[:mdsw.currentBlock.Length]
			if mdsw.preProcessFilter != nil {
				data = mdsw.preProcessFilter(data)
			}

			fileBytesWritten := 0
			if mdsw.currentBlock.HasFlag(StreamBlockFlagIsCompressed) {
				data, err = mdsw.decomp.DecompressData(data)
				if err != nil {
					return len(p), fmt.Errorf("failed decompressing item data block: %w", err)
				}
			}

			if len(data) != 0 {
				if mdsw.preWriteFilter != nil {
					data = mdsw.preWriteFilter(data)
				}

				if !mdsw.ignoreCurrentItemOutput {
					fileBytesWritten, err = mdsw.currentFile.Write(data)
					if err != nil {
						return len(p), fmt.Errorf("failed writing item data: %w", err)
					}
				}
			}

			mdsw.totalBytesWritten += fileBytesWritten

			// Check if all item data has been received for this file
			if mdsw.currentBlock.HasFlag(StreamBlockFlagLastDataBlock) {
				if !mdsw.ignoreCurrentItemOutput {
					// Close the current file
					err = mdsw.currentFile.Close()
					if err != nil {
						return len(p), fmt.Errorf("failed closing item file: %w", err)
					}

					err = mdsw.applyItemAttributes()
					if err != nil {
						return len(p), fmt.Errorf("failed to apply item attributes: %w", err)
					}
				}

				// in case we were in an ignore state, clear this flag for the next file
				mdsw.ignoreCurrentItemOutput = false

				// Reset the current file and item header
				mdsw.currentFile = nil
				mdsw.currentItemHeader = nil
			}

			blockLen := mdsw.currentBlock.Length
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

func (mdsw *MultiDirectoryStreamWriter) applyItemAttributes() error {
	if !mdsw.currentItemNode.PropsIncluded {
		// properties were not stored, so no info to update
		return nil
	}

	nodeTime, err := mdsw.currentItemNode.NodeToTime()
	if err == nil {
		err = os.Chtimes(mdsw.currentFilePath, time.Now(), nodeTime)
		if err != nil {
			return err
		}
	}

	err = os.Chmod(mdsw.currentFilePath, os.FileMode(mdsw.currentItemNode.PermBits))
	if err != nil {
		return err
	}

	return nil
}

func (mdsw *MultiDirectoryStreamWriter) processTreeData(tree Tree) error {
	dirNodes := tree.GetDirNodes()
	for _, dirNode := range dirNodes {
		if len(mdsw.includePaths) > 0 {
			parentPathPrefixLower := strings.ToLower(tree.GetParentPathPrefix())
			isIncluded := false
			for _, includePath := range mdsw.includePaths {
				if strings.Contains(parentPathPrefixLower, strings.ToLower(includePath)) {
					isIncluded = true
					break
				}
			}
			if !isIncluded {
				continue
			}
		}

		targetPath := filepath.Join(mdsw.rootPath, tree.GetParentPathPrefix(), dirNode.Path)
		err := helpers.ForcePath(targetPath)
		if err != nil {
			return fmt.Errorf(
				"Unable to create path in output folder. Path: \"%s\".  Error: %w",
				targetPath,
				err)
		}

		nodeTime, err := dirNode.NodeToTime()
		if err == nil {
			_ = os.Chtimes(targetPath, time.Now(), nodeTime)
		}
	}

	return nil
}

func (mdsw *MultiDirectoryStreamWriter) processItemHeader() error {
	// Open the item's file for writing
	// Get the DirNode
	mdsw.currentDirNode = mdsw.currentTree.GetDirNodeByID(int(mdsw.currentItemHeader.DirID))
	if mdsw.currentDirNode == nil {
		return fmt.Errorf("could not locate directory node with DirID: %d, ItemID: %d",
			mdsw.currentItemHeader.DirID,
			mdsw.currentItemHeader.ItemID)
	}

	mdsw.currentItemNode = mdsw.currentTree.GetItemNodeByID(int(mdsw.currentItemHeader.ItemID))
	if mdsw.currentItemNode == nil {
		return fmt.Errorf("could not locate item node with ItemID: %d",
			mdsw.currentItemHeader.ItemID)
	}

	// Here, we check to see if the parent path prefix is a root path referenced in the
	// includePaths list.  If there is an active include list and the parent path prefix is not
	// in it, then we will ignore this item.
	// This was originally added to support backup/restore and the ability to target specific restore
	// paths in backup files.

	parentPathPrefix := strings.ToLower(mdsw.currentTree.GetParentPathPrefix())

	if len(mdsw.includePaths) > 0 {
		isIncluded := false
		for _, includePath := range mdsw.includePaths {
			if strings.Contains(parentPathPrefix, strings.ToLower(includePath)) {
				isIncluded = true
				break
			}
		}

		if !isIncluded {
			mdsw.ignoreCurrentItemOutput = true
			return nil
		}
	}

	mdsw.currentFilePath = filepath.Join(
		mdsw.rootPath,
		mdsw.currentTree.GetParentPathPrefix(),
		mdsw.currentDirNode.Path,
		mdsw.currentItemNode.Name)

	if mdsw.requireConfirm && helpers.FileExists(mdsw.currentFilePath) {
		if mdsw.overwriteDenyAll {
			mdsw.ignoreCurrentItemOutput = true
			return nil
		}

		shouldOverride, err := mdsw.confirmOverwrite(mdsw.currentItemNode.Name, mdsw.currentFilePath)
		if err != nil {
			return fmt.Errorf("failed to confirm overwrite of current file: %w", err)
		}

		if !shouldOverride {
			mdsw.ignoreCurrentItemOutput = true
			return nil
		}
	}

	file, err := os.Create(mdsw.currentFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	mdsw.currentFile = file
	return nil
}

func (mdsw *MultiDirectoryStreamWriter) confirmOverwrite(fileName, filePath string) (bool, error) {
	fmt.Printf("File \"%s\" already exists at \"%s\".\n", fileName, filePath)

	listItems := []helpers.InputListItem{
		{
			Option: "Y",
			Label:  "Yes, overwrite the current file",
		},
		{
			Option: "N",
			Label:  "No, do not overwrite the current file",
		},
		{
			Option: "A",
			Label:  "Yes to this and ALL future requests",
		},
		{
			Option: "O",
			Label:  "No to this and ALL future requests",
		},
		{
			Option: "C",
			Label:  "Cancel and exit the application",
		},
	}

	selection, err := helpers.GetInputFromList("Overwrite the current file?", listItems, "N")
	fmt.Println("")
	if err != nil {
		return false, err
	}

	switch selection {
	case "Y":
		return true, nil
	case "N":
		return false, nil
	case "A":
		mdsw.requireConfirm = false
		return true, nil
	case "O":
		mdsw.overwriteDenyAll = true
		return false, nil
	case "C":
		return false, errors.New("user cancelled")
	}

	return false, errors.New("invalid response on confirm: cancelling")
}
