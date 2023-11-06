// Copyright 2023 The Bumblebee Authors
//
// Use of this source code is governed by an MIT license that is located
// in this project's root folder, and can also be found online at:
//
// https://github.com/thoughtrealm/bumblebee/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package security should have no internal package dependencies.
// NATS and "helpers" package dependencies are ok.
package security

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

// RandomSignatureData provides an obfuscated sig input structure to increase the complexity of
// known text attacks against the signature output
type RandomSignatureData struct {
	RandomInput []byte
	Signature   []byte
}

type KeyInfo struct {
	IsDefault bool
	KeyType   KeyType
	Name      string
	KeyData   []byte
}

type KeyType int8

const (
	KeyTypeSeed KeyType = iota
	KeyTypePublic
)

var KeyTypeToText = map[KeyType]string{
	KeyTypeSeed:   "SEED",
	KeyTypePublic: "PUBLIC",
}

func (k KeyType) String() string {
	if k != KeyTypeSeed && k != KeyTypePublic {
		return "unknown"
	}

	return KeyTypeToText[k]
}

func NewKey() *KeyInfo {
	return &KeyInfo{}
}

func NewKeyWithKeypair() (*KeyInfo, error) {
	newKeyInfo := NewKey()
	err := newKeyInfo.InitKey()
	return newKeyInfo, err
}

func NewKeyInfo(isDefault bool, keyType KeyType, name string, keyData []byte) (*KeyInfo, error) {
	if keyType != KeyTypeSeed && keyType != KeyTypePublic {
		return nil, fmt.Errorf("unknown key type: %d", int8(keyType))
	}

	if keyData == nil {
		return nil, errors.New("empty key data provided")
	}

	return &KeyInfo{
		IsDefault: isDefault,
		KeyType:   keyType,
		Name:      name,
		KeyData:   keyData,
	}, nil
}

func (ki *KeyInfo) Clone() (keyOut *KeyInfo) {
	keyOut = &KeyInfo{
		IsDefault: ki.IsDefault,
		KeyType:   ki.KeyType,
		Name:      ki.Name,
		KeyData:   bytes.Clone(ki.KeyData),
	}

	return keyOut
}

func (ki *KeyInfo) InitKey() error {
	kp, err := nkeys.CreateCurveKeys()
	if err != nil {
		return err
	}

	ki.KeyType = KeyTypeSeed
	ki.KeyData, err = kp.Seed()
	if err != nil {
		return err
	}

	return nil
}

// CopyFrom will update the current keyinfo with values from the source KeyInfo.
// It returns a KeyInfo ref to support constructions like "return keyOut.CopyFrom()".
func (ki *KeyInfo) CopyFrom(sourceKeyInfo *KeyInfo) *KeyInfo {
	ki.IsDefault = sourceKeyInfo.IsDefault
	ki.KeyType = sourceKeyInfo.KeyType
	ki.Name = sourceKeyInfo.Name
	ki.KeyData = bytes.Clone(sourceKeyInfo.KeyData)
	return ki
}

func (ki *KeyInfo) PublicKey() ([]byte, error) {
	if ki.KeyType == KeyTypePublic {
		return ki.KeyData, nil
	}

	if ki.KeyType != KeyTypeSeed {
		return nil, errors.New("invalid key type detected: expected type PUBLIC or SEED")
	}

	kp, err := nkeys.FromSeed(ki.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed extracting keypair: %w", err)
	}

	publicKey, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed extracting public key from seed: %w", err)
	}

	return []byte(publicKey), nil
}

func (ki *KeyInfo) PrivateKey() ([]byte, error) {
	if ki.KeyType == KeyTypePublic {
		return nil, errors.New("incorrect key type: PUBLIC.  Expected type SEED.")
	}

	if ki.KeyType != KeyTypeSeed {
		return nil, errors.New("invalid key type detected: expected type SEED")
	}

	kp, err := nkeys.FromSeed(ki.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed extracting keypair: %w", err)
	}

	privateKey, err := kp.PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed extracting private key from seed: %w", err)
	}

	return privateKey, nil
}

func (ki *KeyInfo) SignRandom(input []byte) ([]byte, error) {
	salt, err := helpers.GetRandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve random bytes for random sig salt: %w", err)
	}

	signedBytes, err := ki.Sign(salt)
	if err != nil {
		return nil, fmt.Errorf("error signing random data: %w", err)
	}

	rsd := &RandomSignatureData{
		RandomInput: salt,
		Signature:   signedBytes,
	}

	randomSignatureDataBytes, err := msgpack.Marshal(rsd)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall random sig data: %w", err)
	}

	return randomSignatureDataBytes, nil
}

func (ki *KeyInfo) Sign(input []byte) ([]byte, error) {
	if ki.KeyType == KeyTypePublic {
		return nil, errors.New("incorrect key type: PUBLIC:  expected type SEED.")
	}

	if ki.KeyType != KeyTypeSeed {
		return nil, errors.New("invalid key type detected: expected type SEED")
	}

	kp, err := nkeys.FromSeed(ki.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed extracting keypair: %w", err)
	}

	signedBytes, err := kp.Sign(input)
	if err != nil {
		return nil, fmt.Errorf("error signing data: %w", err)
	}

	return signedBytes, nil
}

func (ki *KeyInfo) Verify(input, sig []byte) (isValid bool, err error) {
	var publicKey []byte

	switch ki.KeyType {
	case KeyTypePublic:
		publicKey = ki.KeyData
	case KeyTypeSeed:
		publicKey, err = ki.getPublicKeyFromSeed(ki.KeyData)
		if err != nil {
			return false, fmt.Errorf("error extracting public key from seed: %w", err)
		}
	default:
		return false, errors.New("unknown key type detected in KeyInfo: expected type SEED or PUBLIC")
	}

	kp, err := nkeys.FromPublicKey(string(publicKey))
	if err != nil {
		return false, fmt.Errorf("unable to extract keypair from public key: %w", err)
	}

	err = kp.Verify(input, sig)
	if err != nil {
		return false, fmt.Errorf("")
	}

	return true, nil
}

func (ki *KeyInfo) getPublicKeyFromSeed(seed []byte) ([]byte, error) {
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, fmt.Errorf("unable to extract kp from seed: %w", err)
	}

	publicKey, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get public key from keypair: %w", err)
	}

	return []byte(publicKey), nil
}

// VerifyRandomSignature will decode the RandomSignatureData struct and call Verify on the internal constructs
func (ki *KeyInfo) VerifyRandomSignature(input []byte) (isValid bool, err error) {
	rsd := &RandomSignatureData{}
	err = msgpack.Unmarshal(input, rsd)
	if err != nil {
		return false, fmt.Errorf("unable to extract RandomSignatureData from input: %w", err)
	}

	return ki.Verify(rsd.RandomInput, rsd.Signature)
}

func (ki *KeyInfo) Wipe() {
	Wipe(ki.KeyData)
}

func Wipe(refBytes []byte) {
	if len(refBytes) != 0 {
		_, _ = io.ReadFull(rand.Reader, refBytes[:])
	}
}
