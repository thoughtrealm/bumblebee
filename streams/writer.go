package streams

type StreamWriter interface {
}

type MultiDirectoryStreamWriter struct {
	decomp Decompressor
	Trees  []Tree
}

func NewMultiDirectoryStreamWriter() (StreamWriter, error) {
	decomp, err := NewDecompressor()
	if err != nil {
		return nil, err
	}

	return &MultiDirectoryStreamWriter{
		decomp: decomp,
		Trees:  make([]Tree, 0),
	}, nil
}
