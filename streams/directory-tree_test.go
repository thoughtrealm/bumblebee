package streams

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const TREE_TEST_PATH = "testdir"

func TestNewDirectoryTree(t *testing.T) {
	dt := NewDirectoryTree(
		WithDirectoryIncludePatterns(
			[]string{"*.*"},
		),
	)

	assert.NotNil(t, dt)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 0, dirCount)
	assert.Equal(t, 0, itemCount)

	printDirsHelper(t, dt)
}

func TestNewDirectoryTreeFromPathWithEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH, WithEmptyPaths())

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 7, dirCount)
	assert.Equal(t, 3, itemCount)

	printDirsHelper(t, dt)
}

func TestNewDirectoryTreeFromPathNoEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH)

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 6, dirCount)
	assert.Equal(t, 3, itemCount)

	printDirsHelper(t, dt)
}

func TestNewDirectoryTreeFromPathIncludedDirOnly(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH,
		WithDirectoryIncludePatterns([]string{"dir1", "dir1-2"}),
	)

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 3, dirCount)
	assert.Equal(t, 1, itemCount)

	printDirsHelper(t, dt)
}

func TestNewDirectoryTreeFromPathWithPathExcludesAndNoEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH,
		WithDirectoryExcludePatterns([]string{"dir1"}),
	)

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 2, dirCount)
	assert.Equal(t, 1, itemCount)

	printDirsHelper(t, dt)
}

func TestNewDirectoryTreeFromPathWithPathIncludesAndExcludesAndEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH,
		WithEmptyPaths(),
		WithDirectoryIncludePatterns([]string{"dir1", "Dir2", "empty-dir"}),
		WithDirectoryExcludePatterns([]string{"dir1"}),
	)

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount, _ := dt.Stats()
	assert.Equal(t, 3, dirCount)
	assert.Equal(t, 1, itemCount)

	printDirsHelper(t, dt)
}

func TestDirectoryTree_ListDirs(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath(TREE_TEST_PATH, WithEmptyPaths())
	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirs := dt.ListDirs(true)
	assert.Equal(t, []string{
		"/",
		"/dir1",
		"/dir1/dir1-2",
		"/dir1/dironly",
		"/dir1/dironly/dir2-1",
		"/Dir2",
		"/Dir2/empty-dir",
	}, dirs)

	dirs = dt.ListDirs(false)
	assert.Equal(t, []string{
		"/",
		"/Dir2",
		"/Dir2/empty-dir",
		"/dir1",
		"/dir1/dir1-2",
		"/dir1/dironly",
		"/dir1/dironly/dir2-1",
	}, dirs)
}

func TestDirectoryTree_ToAndFromBytes(t *testing.T) {
	dtSourceTree, err := NewDirectoryTreeFromPath(TREE_TEST_PATH, WithEmptyPaths(), WithItemDetails())
	assert.NotNil(t, dtSourceTree)
	assert.Nil(t, err)

	dirCount, itemCount, totalBytes := dtSourceTree.Stats()
	t.Logf("dirCount  : %d", dirCount)
	t.Logf("itemCount : %d", itemCount)
	t.Logf("totalBytes: %d", totalBytes)

	toBytes, err := dtSourceTree.ToBytes()
	assert.NotNil(t, toBytes)
	assert.Nil(t, err)

	t.Logf("len(toBytes): %d", len(toBytes))
	dtFromTree := NewDirectoryTree(WithEmptyPaths())
	err = dtFromTree.FromBytes(toBytes)
	assert.Nil(t, err)

	// Get the concrete type so we can tell if the props on the directory are the same as the source tree
	dtFromTreeConcrete, isValid := dtFromTree.(*DirectoryTree)
	assert.True(t, isValid)
	assert.IsType(t, dtFromTreeConcrete, &DirectoryTree{})
	assert.True(t, dtFromTreeConcrete.IncludeItemDetails)

	sourceDirNodes := dtSourceTree.GetDirNodes()
	fromDirNodes := dtFromTree.GetDirNodes()

	sourceItemNodes := dtSourceTree.GetItemNodes()
	fromItemNodes := dtFromTree.GetItemNodes()

	for idx, sourceDirNode := range sourceDirNodes {
		assert.True(t, sourceDirNode.Compare(fromDirNodes[idx]))
	}

	for idx, sourceItemNode := range sourceItemNodes {
		assert.True(t, sourceItemNode.Compare(fromItemNodes[idx]))
	}
}

func printDirsHelper(t *testing.T, dt Tree) {
	dirs := dt.ListDirs(true)
	outputText := "\nDirectory List\n"
	outputText += "=============================\n"
	for _, d := range dirs {
		outputText += fmt.Sprintf("\"%s\"\n", d)
	}

	t.Log(outputText)
}
