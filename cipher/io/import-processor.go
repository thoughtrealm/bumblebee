package io

import (
	"errors"
	"github.com/thoughtrealm/bumblebee/security"
)

type ProcessorAcquirePasswordFunc func() (password []byte, err error)

type ImportProcessor struct {
	password            []byte
	acquirePasswordFunc ProcessorAcquirePasswordFunc
	importedUser        *security.KeyInfo
	importedKeyPair     *security.KeyPairInfo
	importDataType      security.ExportDataType
}

func NewImportProcessor(passwordFunc ProcessorAcquirePasswordFunc) *ImportProcessor {
	return &ImportProcessor{acquirePasswordFunc: passwordFunc}
}

func (ip *ImportProcessor) DataType() security.ExportDataType {
	return ip.importDataType
}

func (ip *ImportProcessor) ImportedUser() *security.KeyInfo {
	return ip.importedUser
}

func (ip *ImportProcessor) ImportedKeyPair() *security.KeyPairInfo {
	return ip.importedKeyPair
}

func (ip *ImportProcessor) Wipe() {
	security.Wipe(ip.password)
}

func (ip *ImportProcessor) ProcessImportData(data []byte) (err error) {
	defer ip.Wipe()
	// todo: left off here... implement this func
	return errors.New("ProcessImportData not implemented")
}
