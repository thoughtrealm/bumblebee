package streams

type StreamWriter interface {
}

type MultiDirectoryStreamWriter struct {
}

func newMultiDirectoryStreamWriter(dirs []string) (StreamWriter, error) {
	return &MultiDirectoryStreamWriter{}, nil
}
