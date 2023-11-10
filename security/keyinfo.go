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
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"github.com/vmihailenco/msgpack/v5"
)

// RandomSignatureData provides an obfuscated sig input structure to increase the complexity of
// known text attacks against the signature output
type RandomSignatureData struct {
	RandomInput []byte
	Signature   []byte
}

type KeyInfo struct {
	Name          string
	CipherPubKey  string
	SigningPubKey string
}

func NewKey() *KeyInfo {
	return &KeyInfo{}
}

func NewKeyInfo(name string, cipherPubKey, signingPubKey string) (*KeyInfo, error) {
	if cipherPubKey == "" {
		return nil, errors.New("empty cipher key data provided")
	}

	if signingPubKey == "" {
		return nil, errors.New("empty signing key data provided")
	}

	return &KeyInfo{
		Name:          name,
		CipherPubKey:  cipherPubKey,
		SigningPubKey: signingPubKey,
	}, nil
}

func (ki *KeyInfo) Clone() (keyOut *KeyInfo) {
	keyOut = &KeyInfo{
		Name:          ki.Name,
		CipherPubKey:  ki.CipherPubKey,
		SigningPubKey: ki.SigningPubKey,
	}

	return keyOut
}

// CopyFrom will update the current keyinfo with values from the source KeyInfo.
// It returns a KeyInfo ref to support constructions like "return keyOut.CopyFrom()".
func (ki *KeyInfo) CopyFrom(sourceKeyInfo *KeyInfo) *KeyInfo {
	ki.Name = sourceKeyInfo.Name
	ki.CipherPubKey = sourceKeyInfo.CipherPubKey
	ki.SigningPubKey = sourceKeyInfo.SigningPubKey
	return ki
}

func (ki *KeyInfo) IsSameData(sourceKeyInfo *KeyInfo) bool {
	if ki.Name == sourceKeyInfo.Name &&
		ki.CipherPubKey == sourceKeyInfo.CipherPubKey &&
		ki.SigningPubKey == sourceKeyInfo.SigningPubKey {
		return true
	}

	return false
}

func (ki *KeyInfo) Verify(input, sig []byte) (isValid bool, err error) {
	verifyKP, err := nkeys.FromPublicKey(ki.SigningPubKey)
	if err != nil {
		return false, fmt.Errorf("unable to extract verify keypair from public key: %w", err)
	}

	err = verifyKP.Verify(input, sig)
	if err != nil {
		return false, fmt.Errorf("")
	}

	return true, nil
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
