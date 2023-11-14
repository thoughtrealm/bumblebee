package io

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/cipher"
	"github.com/thoughtrealm/bumblebee/security"
	"io"
	"os"
)

type ExportWriter struct {
	password []byte
}

func NewExportWriter(password []byte) *ExportWriter {
	return &ExportWriter{password: bytes.Clone(password)}
}

func (ew *ExportWriter) Wipe() {
	security.Wipe(ew.password)
}

// WriteExportKeyInfoToFile writes security.ExportKeyInfo data to a file.  It will simply open the file,
// then pass it to WriteKeyInfoToStream().
func (ew *ExportWriter) WriteExportKeyInfoToFile(eki *security.ExportKeyInfo, password []byte, filePath string) error {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	err = ew.WriteExportKeyInfoToStream(eki, password, outputFile)
	if err != nil {
		return fmt.Errorf("unable to write export info to file: %w", err)
	}

	return nil
}

// WriteExportKeyInfoToStream writes security.ExportKeyInfo data to a stream writer.  If the eki.InfoType is
// KeyInfo, then the password is optional, since that only contains public keys.  If it is
// type KeyPairInfo, then the password is required and an error is returned if it is nil.
func (ew *ExportWriter) WriteExportKeyInfoToStream(eki *security.ExportKeyInfo, password []byte, w io.Writer) error {
	if eki.DataType == security.ExportKeyInfoTypeKeyPairInfo && len(password) == 0 {
		return errors.New("cannot export: no password provided and password is required for exporting keypair data")
	}

	ekiBytes, err := ew.ekiToEncryptedBytes(eki, password)
	if err != nil {
		return fmt.Errorf("unable to serialize ExportKeyInfo for stream output: %w", err)
	}

	_, err = w.Write(ekiBytes)
	if err != nil {
		return fmt.Errorf("unable to write ExportKeyInfo to output stream: %w", err)
	}

	return nil
}

func (ew *ExportWriter) ekiToEncryptedBytes(eki *security.ExportKeyInfo, password []byte) (ekiBytesOut []byte, err error) {
	ekiBytes, err := eki.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed encoding ExportKeyInfo to bytes: %w", err)
	}

	if len(password) == 0 {
		zeroByteLen := []byte{0}
		return append(zeroByteLen, ekiBytes...), nil
	}

	ekiBytesEncrypted, salt, err := cipher.EncryptBytes(ekiBytes, password)
	if err != nil {
		return nil, fmt.Errorf("failed encrypting export output: %w", err)
	}

	ekiBuffer := bytes.NewBuffer(nil)
	_, err = WriteBytesTo(salt, LenMarkerSize8, ekiBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed writing salt to encrypted stream: %w", err)
	}

	_, err = ekiBuffer.Write(ekiBytesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed writing encrypted data to stream: %w", err)
	}

	return ekiBuffer.Bytes(), nil
}

func BytesToEKI(ekiBytes []byte, password []byte) (eki *security.ExportKeyInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered in BytesToExportKeyInfo: %s", r)
		}
	}()

	if len(ekiBytes) == 0 {
		return nil, errors.New("input is empty")
	}

	var useBytes []byte
	saltLen := int(ekiBytes[0])
	if saltLen != 0 {
		// this is encrypted... grab the salt at the front and decrypt the body
		salt := ekiBytes[1 : saltLen+1]
		useBytes, err = cipher.DecryptBytes(ekiBytes[saltLen+1:], password, salt)
		if err != nil {
			return nil, fmt.Errorf("error attempting to decrypt the export key info: %w", err)
		}
	} else {
		useBytes = ekiBytes[1:]
	}

	eki, err = security.NewExportKeyInfoFromBytes(useBytes)
	if err != nil {
		return nil, fmt.Errorf("failed reading export key info bytes: %w", err)
	}

	return eki, err
}
