package symfiles

import "errors"

func (ssf *SimpleSymFile) WriteSymFileFromFile(key []byte, inputFile, outputSymFile string) (bytesWritten int64, bytesRead int64, err error) {
	return 0, 0, errors.New("WriteSymFileFromFile not implemented")
}

func (ssf *SimpleSymFile) WriteSymFileFromDirs(key []byte, inputDirs []string, outputSymFile string) (bytesWritten int64, bytesRead int64, err error) {
	return 0, 0, errors.New("WriteSymFileFromDirs not implemented")
}
