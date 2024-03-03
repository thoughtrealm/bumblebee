package symfiles

import (
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
)

type SymFile interface {
	ReadSymFile(key []byte, inputSymFilename, outputPath string) (bytesWritten int, err error)
	WriteSymFileFromFile(key []byte, inputFilename, outputSymFileName string) (bytesWritten int, err error)
	WriteSymFileFromDirs(key []byte, inputDirs []string, outputSymFileName string) (bytesWritten int, err error)
}

type SimpleSymFile struct {
	sc beecipher.Cipher
}

func NewSymFile(key []byte) (SymFile, error) {
	newCipher, err := beecipher.NewSymmetricCipher(key, DEFAULT_CHUNK_SIZE)
	if err != nil {
		return nil, err
	}

	return &SimpleSymFile{
		sc: newCipher,
	}, nil
}
