package symfiles

import "errors"

// ReadSymFileToFile reads a .bsym file from the inputSymFilePath and writes it to the outputFile.
// It returns the number of bytes read, the number of bytes written, and any error encountered.
func (ssf *SimpleSymFile) ReadSymFileToFile(inputSymFile, outputFile string) (bytesRead, bytesWritten int64, err error) {
	return 0, 0, errors.New("ReadSymFileToFile not implemented")
}

// ReadSymFileToPath reads a .bsym file with multi-dir data from the inputSymFilePath and writes it to the outputPath.
// It returns the number of bytes read, the number of bytes written, and any error encountered.
func (ssf *SimpleSymFile) ReadSymFileToPath(inputSymFile, outputPath string) (bytesRead, bytesWritten int64, err error) {
	return 0, 0, errors.New("ReadSymFileToFile not implemented")
}
