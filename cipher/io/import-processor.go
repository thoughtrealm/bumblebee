package io

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/vmihailenco/msgpack/v5"
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

	if len(data) == 0 {
		return errors.New("import data is nil")
	}

	var useBytes []byte
	if data[0] != 0 {
		// this indicates that a password is required
		useBytes, err = ip.decryptData(data)
		if err != nil {
			return err
		}
	} else {
		useBytes = data[1:]
	}

	var eki = &security.ExportKeyInfo{}
	err = msgpack.Unmarshal(useBytes, eki)
	if err != nil {
		return fmt.Errorf("Unable to read import data structure: %w", err)
	}

	// see what kind of import it is
	ip.importDataType = eki.DataType
	switch eki.DataType {
	case security.ExportDataTypeKeyInfo:
		ip.importedUser, _ = security.NewKeyInfo(eki.Name, eki.CipherPubKey, eki.SigningPubKey)
	case security.ExportDataTypeKeyPairInfo:
		ip.importedKeyPair = security.NewKeyPairInfoFromSeeds(eki.Name, eki.CipherSeed, eki.SigningSeed)
	case security.ExportDataTypeUnknown:
		return errors.New("imported data has a date type of UNKNOWN")
	default:
		return fmt.Errorf("unknown imported data type ID: %d", int(eki.DataType))
	}

	return nil
}

func (ip *ImportProcessor) decryptData(data []byte) (decryptedData []byte, err error) {
	// need to get a key first
	if ip.acquirePasswordFunc == nil {
		return nil, errors.New("a password is required, but no mechanism is provided to obtain one")
	}

	password, err := ip.acquirePasswordFunc()
	if err != nil {
		return nil, err
	}

	// we already know a salt exists, so no need to confirm salt byte marker
	saltLen := data[0]
	salt := data[1 : saltLen+1]
	encryptedBytes := data[saltLen+1:]

	return cipher.DecryptBytes(encryptedBytes, password, salt)
}
