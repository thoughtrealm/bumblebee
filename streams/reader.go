package streams

import "fmt"

type StreamReader interface {
	AddDir(dir string) (newTree Tree, err error)
}

type MultiDirectoryStreamReader struct {
	Comp  Compressor
	Trees []Tree
}

func NewMultiDirectoryStreamReader() (StreamReader, error) {
	comp, err := NewCompressor()
	if err != nil {
		return nil, err
	}

	return &MultiDirectoryStreamReader{
		Comp:  comp,
		Trees: make([]Tree, 0),
	}, nil
}

func NewMultiDirectoryStreamReaderFromDirs(dirs []string) (StreamReader, error) {
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no directories provided")
	}

	mdsr, err := NewMultiDirectoryStreamReader()
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		_, err = mdsr.AddDir(dir)
		if err != nil {
			return nil, err
		}
	}

	return mdsr, nil
}

func (mdsr *MultiDirectoryStreamReader) AddDir(dir string) (newTree Tree, err error) {
	return nil, nil
}
