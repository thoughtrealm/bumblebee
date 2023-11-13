package security

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
)

type ExportKeyInfoType int

const (
	// ExportKeyInfoTypeKeyInfo will be persisted to export files, so we will use explicit values instead
	// of iota.  If we add types in the future or change things around, we don't want to deprecate or
	// invalidate exported files.
	ExportKeyInfoTypeKeyInfo     ExportKeyInfoType = 1
	ExportKeyInfoTypeKeyPairInfo ExportKeyInfoType = 2
)

type ExportKeyInfo struct {
	Name          string
	InfoType      ExportKeyInfoType
	CipherSeed    []byte
	SigningSeed   []byte
	CipherPubKey  string
	SigningPubKey string
}

func NewExportKeyInfo() *ExportKeyInfo {
	return &ExportKeyInfo{}
}

func NewExportKeyInfoFromKeyPairInfo(kpi *KeyPairInfo) (*ExportKeyInfo, error) {
	if kpi == nil {
		return nil, errors.New("keypair info input is nil")
	}

	cipherPubKey, signingPubKey, err := kpi.PublicKeys()
	if err != nil {
		return nil, fmt.Errorf("unable to extract public keys from keypair data: %w", err)
	}

	return &ExportKeyInfo{
		Name:          kpi.Name,
		InfoType:      ExportKeyInfoTypeKeyPairInfo,
		CipherSeed:    kpi.CipherSeed,
		SigningSeed:   kpi.SigningSeed,
		CipherPubKey:  cipherPubKey,
		SigningPubKey: signingPubKey,
	}, nil
}

func NewExportKeyInfoFromKeyInfo(ki *KeyInfo) (*ExportKeyInfo, error) {
	if ki == nil {
		return nil, errors.New("keyinfo input is nil")
	}

	return &ExportKeyInfo{
		Name:          ki.Name,
		InfoType:      ExportKeyInfoTypeKeyInfo,
		CipherSeed:    nil,
		SigningSeed:   nil,
		CipherPubKey:  ki.CipherPubKey,
		SigningPubKey: ki.SigningPubKey,
	}, nil
}

func NewExportKeyInfoFromBytes(ekiBytes []byte) (*ExportKeyInfo, error) {
	var eki = &ExportKeyInfo{}
	err := msgpack.Unmarshal(ekiBytes, eki)
	if err != nil {
		return nil, fmt.Errorf("failed decoding ExportKeyInfo bytes: %w", err)
	}

	return eki, nil
}

func (eki *ExportKeyInfo) ToBytes() (exportBytes []byte, err error) {
	return msgpack.Marshal(eki)
}
