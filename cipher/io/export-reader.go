package io

import (
	"errors"
	"github.com/thoughtrealm/bumblebee/security"
)

type ReaderAcquirePasswordFunc func() (password []byte)

type ExportReader struct {
	password            []byte
	acquirePasswordFunc ReaderAcquirePasswordFunc
}

func NewExportReader(password []byte) *ExportReader {
	return &ExportReader{password: password}
}

func NewExportReaderWithAcquireFunc(acquirePasswordFunc ReaderAcquirePasswordFunc) *ExportReader {
	return &ExportReader{acquirePasswordFunc: acquirePasswordFunc}
}

func (er *ExportReader) Wipe() {
	security.Wipe(er.password)
}

func (er *ExportReader) ReadUserInfoFromFile(filePath string) (*security.KeyInfo, error) {
	return nil, errors.New("ReadUserInfoFromFile not implemented")
}

func (er *ExportReader) ReadKeyPairInfoFromFile(filePath string) (*security.KeyInfo, error) {
	return nil, errors.New("ReadKeyPairInfoFromFile not implemented")
}
