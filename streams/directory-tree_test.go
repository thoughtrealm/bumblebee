package streams

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDirectoryTree(t *testing.T) {
	dt := NewDirectoryTree(
		WithDirectoryIncludePattern(
			[]string{"*.*"},
		),
	)

	assert.NotNil(t, dt)

	dirCount, itemCount := dt.Stats()
	assert.Equal(t, 0, dirCount)
	assert.Equal(t, 0, itemCount)
}

func TestNewDirectoryTreeFromPathWithEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath("testdir", WithEmptyPaths())

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount := dt.Stats()
	assert.Equal(t, 3, dirCount)
	assert.Equal(t, 1, itemCount)
}

func TestNewDirectoryTreeFromPathWithoutEmptyPaths(t *testing.T) {
	dt, err := NewDirectoryTreeFromPath("testdir")

	assert.NotNil(t, dt)
	assert.Nil(t, err)

	dirCount, itemCount := dt.Stats()
	assert.Equal(t, 2, dirCount)
	assert.Equal(t, 1, itemCount)
}
