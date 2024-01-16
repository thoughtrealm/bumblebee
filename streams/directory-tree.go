package streams

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/helpers"
	"os"
	"path/filepath"
	"slices"
	"time"
)

/*
	A directory tree is a structure for storing paths from a starting root path.

	Each path is stored as a node with a unique ID and all the properties of the path.

	Each node can be mapped to files.

	To prevent name conflicts, we refer to files as a ItemNode.

	Each ItemNode has a unique ID and all the relevant properties of the file.
*/

type ItemNode struct {
	DirID         int
	ItemID        int
	Name          string
	NodeTime      time.Time
	OriginalSize  int64
	PermBits      uint16
	PropsIncluded bool
}

type DirNode struct {
	DirID         int
	Name          string
	Path          string
	NodeTime      time.Time
	PermBits      uint16
	PropsIncluded bool
}

type Tree interface {
	FromBytes() error
	ScanPath(rootPath string) error
	Stats() (dirCount, itemCount int)
	ToBytes() ([]byte, error)
}

type DirectoryTree struct {
	// IncludeEmptyPaths will add paths with no files in them to the directory tree.
	// Paths "." and ".." are always ignored, regardless of this setting.
	IncludeEmptyPaths bool

	// IncludeItemDetails will populate the ItemNode's detail fields.  When the data is extracted,
	// Bumblebee will attempt to apply the detail values to the created file.
	// When IncludeItemDetails is not set, only the ID, NodeID and Name are set.  All other values are
	// left as default values when creating the item during extraction.
	IncludeItemDetails bool

	// DirectorySearchPattern will be used to determine the directories to include when scanning the root path.
	// If empty, all directories will be added.
	DirectoryIncludePatterns []string

	// DirectoryExcludePatterns will be used to exclude directories based on the provided patterns.
	// If empty, no directories will be excluded.
	DirectoryExcludePatterns []string

	// FileIncludePatterns will be used to determine the files to include when scanning the root path.
	// If empty, all files will be included.
	FileIncludePatterns []string

	// FileExcludePatterns will be used to exclude files based on the provided patterns.
	// If empty, no files will be excluded
	FileExcludePatterns []string

	// RootPath is used during the root scan to determine relative pathing from the base root.
	RootPath string

	// DirNodes will contain a DirNode for each included directory path
	DirNodes []*DirNode

	// ItemNodes will contain an ItemNode for every file included in the tree
	ItemNodes []*ItemNode `msgpack:"-"`

	// NextDirID identifies the DirNode and will be incremented for each DirNode
	NextDirID int

	// NextItemID identifies the ItemNode and will be incremented for each ItemNode
	NextItemID int
}

type DirectoryOption func(tree *DirectoryTree)

func WithItemDetails() DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.IncludeItemDetails = true
	}
}

func WithEmptyPaths() DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.IncludeEmptyPaths = true
	}
}

func WithFileIncludePattern(includePatterns []string) DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.FileIncludePatterns = slices.Clone(includePatterns)
	}
}

func WithFileExcludePattern(excludePatterns []string) DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.FileExcludePatterns = slices.Clone(excludePatterns)
	}
}

func WithDirectoryIncludePattern(includePatterns []string) DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.DirectoryIncludePatterns = slices.Clone(includePatterns)
	}
}

func WithDirectoryExcludePattern(excludePatterns []string) DirectoryOption {
	return func(dt *DirectoryTree) {
		dt.DirectoryExcludePatterns = slices.Clone(excludePatterns)
	}
}

func NewDirectoryTree(options ...DirectoryOption) Tree {
	dt := &DirectoryTree{}

	for _, option := range options {
		option(dt)
	}

	return dt
}

func NewDirectoryTreeFromPath(rootPath string, options ...DirectoryOption) (Tree, error) {
	newDirTree := NewDirectoryTree(options...)
	err := newDirTree.ScanPath(rootPath)
	if err != nil {
		return nil, err
	}

	return newDirTree, nil
}

func (dt *DirectoryTree) Clear() {
	dt.DirNodes = []*DirNode{}
	dt.ItemNodes = []*ItemNode{}
	dt.RootPath = ""
	dt.NextDirID = 0
	dt.NextItemID = 0
}

func (dt *DirectoryTree) ScanPath(rootPath string) error {
	pathToScan := helpers.RemoveTrailingPathSeparator(rootPath)
	found, isDir := helpers.PathExistsInfo(pathToScan)
	if !found {
		return fmt.Errorf("root path not found: \"%s\"", pathToScan)
	}

	if !isDir {
		return fmt.Errorf("root path is a file, not a directory")
	}

	dt.Clear()
	dt.RootPath = pathToScan

	_, err := dt.doScanPath(pathToScan)
	return err
}

func (dt *DirectoryTree) doScanPath(fullPath string) (itemsAdded bool, err error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed obtaining path info: %w", err)
	}

	relativePath := dt.getRelativePath(fullPath)
	// We always include the root path, regardless of what is passed for include/exclude patterns
	// So we make sure the relative path is not "", which would be the root path
	if relativePath != "" {
		var includeDir bool
		includeDir, err = dt.shouldIncludeDirectory(info.Name())
		if err != nil {
			return false, fmt.Errorf("failed validating path inclusion: %w", err)
		}

		if !includeDir {
			// Ignore this directory and any subs or files if it is not included
			return false, nil
		}
	}

	thisDirNode := &DirNode{
		DirID:         dt.GetNextDirID(),
		Name:          "",
		Path:          relativePath,
		NodeTime:      time.Time{},
		PermBits:      uint16(uint32(info.Mode()) & uint32(0x1FF)),
		PropsIncluded: false,
	}

	itemsHaveBeenAdded := false
	defer func() {
		if err != nil {
			return
		}

		if itemsHaveBeenAdded || dt.IncludeEmptyPaths {
			dt.DirNodes = append(dt.DirNodes, thisDirNode)
		}
	}()

	// Scan for items in this path
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed reading directory entries for dir \"%s\": %w", fullPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// This is a directory
			// The doScanPath call will check if the directory should be included or not, so we don't check here
			dirPath := filepath.Join(fullPath, entry.Name())
			itemsAdded, err = dt.doScanPath(dirPath)
			if err != nil {
				return false, err
			}

			if itemsAdded {
				itemsHaveBeenAdded = true
			}

			continue
		}

		// This is a file...
		var shouldIncludeFile bool
		shouldIncludeFile, err = dt.shouldIncludeFile(entry.Name())
		if err != nil {
			return false, fmt.Errorf("failed ")
		}

		if !shouldIncludeFile {
			continue
		}

		err = dt.addFileNode(thisDirNode, entry)
		if err != nil {
			return false, fmt.Errorf("failed adding ItemNode for file at \"%s\": %w", fullPath, err)
		}

		itemsHaveBeenAdded = true
	}

	return itemsHaveBeenAdded, nil
}

func (dt *DirectoryTree) addFileNode(dirNode *DirNode, entry os.DirEntry) error {
	thisItemNode := &ItemNode{
		DirID:  dirNode.DirID,
		ItemID: dt.GetNextItemID(),
		Name:   entry.Name(),
	}

	if dt.IncludeItemDetails {
		fileInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed attempting to retrieve info on DirEntry: %w", err)
		}

		thisItemNode.NodeTime = time.Time{}
		thisItemNode.OriginalSize = fileInfo.Size()
		thisItemNode.PermBits = uint16(uint32(fileInfo.Mode()) & uint32(0x1FF))
		thisItemNode.PropsIncluded = true
	}

	dt.ItemNodes = append(dt.ItemNodes, thisItemNode)
	return nil
}

func (dt *DirectoryTree) shouldIncludeDirectory(name string) (shouldInclude bool, err error) {
	if name == "." || name == "..." {
		return false, nil
	}

	if len(dt.DirectoryIncludePatterns) == 0 && len(dt.DirectoryExcludePatterns) == 0 {
		return true, nil
	}

	for _, excludePattern := range dt.DirectoryExcludePatterns {
		matches, err := filepath.Match(excludePattern, name)
		if err != nil {
			return false, err
		}

		if matches {
			return false, nil
		}
	}

	if len(dt.DirectoryIncludePatterns) == 0 {
		// no include pattern provided, so assume everything should be included
		return true, nil
	}

	// Include patterns were provided, so at least one pattern should match the name in order to include the directory
	for _, includePattern := range dt.DirectoryIncludePatterns {
		matches, err := filepath.Match(includePattern, name)
		if err != nil {
			return false, err
		}

		if matches {
			return true, nil
		}
	}

	// Include patterns exist and no include pattern was matched, so should not be included
	return false, nil
}

func (dt *DirectoryTree) shouldIncludeFile(name string) (shouldInclude bool, err error) {
	if len(dt.FileIncludePatterns) == 0 && len(dt.FileExcludePatterns) == 0 {
		return true, nil
	}

	for _, excludePattern := range dt.FileExcludePatterns {
		matches, err := filepath.Match(excludePattern, name)
		if err != nil {
			return false, err
		}

		if matches {
			return false, nil
		}
	}

	if len(dt.FileIncludePatterns) == 0 {
		// no include pattern provided, so assume everything should be included
		return true, nil
	}

	// Include patterns were provided, so at least one pattern should match the name in order to include the directory
	for _, includePattern := range dt.FileIncludePatterns {
		matches, err := filepath.Match(includePattern, name)
		if err != nil {
			return false, fmt.Errorf("failed matching file: %w", err)
		}

		if matches {
			return true, nil
		}
	}

	// Include patterns exist and no include pattern was matched, so should not be included
	return false, nil
}

func (dt *DirectoryTree) GetNextDirID() int {
	dt.NextDirID += 1
	return dt.NextDirID
}

func (dt *DirectoryTree) GetNextItemID() int {
	dt.NextItemID += 1
	return dt.NextItemID
}

func (dt *DirectoryTree) getRelativePath(aOffsetPath string) string {
	aOffsetPathFixed := helpers.RemoveTrailingPathSeparator(aOffsetPath)
	if len(dt.RootPath) == len(aOffsetPathFixed) {
		return ""
	}

	if len(dt.RootPath) > len(aOffsetPathFixed) {
		// This should never happen... panic?  something else?  For now, return empty path ref?
		return ""
	}

	return aOffsetPathFixed[len(dt.RootPath):]
}

func (dt *DirectoryTree) ToBytes() ([]byte, error) {
	return nil, errors.New("ToBytes is not implemented")
}

func (dt *DirectoryTree) FromBytes() error {
	return errors.New("FromBytes is not implemented")
}

func (dt *DirectoryTree) Stats() (dirCount, itemCount int) {
	return len(dt.DirNodes), len(dt.ItemNodes)
}
