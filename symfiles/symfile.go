package symfiles

import (
	beecipher "github.com/thoughtrealm/bumblebee/cipher"
)

type SymFile interface{}

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
