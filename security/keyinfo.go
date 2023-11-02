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

package security

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
)

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

func (k *KeyInfo) Clone() (keyOut *KeyInfo) {
	keyOut = &KeyInfo{
		IsDefault: k.IsDefault,
		KeyType:   k.KeyType,
		Name:      k.Name,
		KeyData:   bytes.Clone(k.KeyData),
	}

	return keyOut
}

func (k *KeyInfo) InitKey() error {
	kp, err := nkeys.CreateCurveKeys()
	if err != nil {
		return err
	}

	k.KeyType = KeyTypeSeed
	k.KeyData, err = kp.Seed()
	if err != nil {
		return err
	}

	return nil
}

// CopyFrom will update the current keyinfo with values from the source KeyInfo.
// It returns a KeyInfo ref to support constructions like "return keyOut.CopyFrom()".
func (k *KeyInfo) CopyFrom(sourceKeyInfo *KeyInfo) *KeyInfo {
	k.IsDefault = sourceKeyInfo.IsDefault
	k.KeyType = sourceKeyInfo.KeyType
	k.Name = sourceKeyInfo.Name
	k.KeyData = bytes.Clone(sourceKeyInfo.KeyData)
	return k
}

func (k *KeyInfo) PublicKey() ([]byte, error) {
	if k.KeyType == KeyTypePublic {
		return k.KeyData, nil
	}

	if k.KeyType != KeyTypeSeed {
		return nil, errors.New("invalid key type detected: expected type PUBLIC or SEED")
	}

	kp, err := nkeys.FromSeed(k.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed extracting keypair: %w", err)
	}

	publicKey, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed extracting public key from seed: %w", err)
	}

	return []byte(publicKey), nil
}

func (k *KeyInfo) PrivateKey() ([]byte, error) {
	if k.KeyType == KeyTypePublic {
		return nil, errors.New("incorrect key type: PUBLIC.  Expected type SEED.")
	}

	if k.KeyType != KeyTypeSeed {
		return nil, errors.New("invalid key type detected: expected type SEED")
	}

	kp, err := nkeys.FromSeed(k.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed extracting keypair: %w", err)
	}

	privateKey, err := kp.PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed extracting private key from seed: %w", err)
	}

	return privateKey, nil
}
