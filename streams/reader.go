package streams

type StreamReader interface {
}

type MultiDirectoryStreamReader struct {
}

func newMultiDirectoryStreamReader(dirs []string) (StreamReader, error) {

	return &MultiDirectoryStreamReader{}, nil
}

func (mdsr *MultiDirectoryStreamReader) AddItems(itemPaths []string) (dirCount, fileCount int, err error) {
	return 0, 0, nil
}
