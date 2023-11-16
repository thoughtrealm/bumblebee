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
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

// KeyPairInfo defines the local environment's cipher keypairs created for sending/receiving and signing/verifying bundles
type KeyPairInfo struct {
	// When initialized, default is the main key.
	// More keys can be added with user-defined names.
	Name string

	// Seed is stored as NATS base32 string
	CipherSeed  []byte
	SigningSeed []byte
}

func NewKeyPairInfoWithSeeds(name string) (*KeyPairInfo, error) {
	// seeds are cloned in the call to NewKeyPairInfoFromSeeds at the end of this func,
	// so it's ok to wipe the following constructs in defers
	cipherKP, err := nkeys.CreateCurveKeys()
	if err != nil {
		return nil, err
	}
	defer cipherKP.Wipe()

	cipherSeed, err := cipherKP.Seed()
	if err != nil {
		return nil, err
	}
	defer Wipe(cipherSeed)

	signingKP, err := nkeys.CreateUser()
	if err != nil {
		return nil, err
	}
	defer signingKP.Wipe()

	signingSeed, err := signingKP.Seed()
	if err != nil {
		return nil, err
	}
	defer Wipe(signingSeed)

	return NewKeyPairInfoFromSeeds(name, cipherSeed, signingSeed), nil
}

func NewKeyPairInfoFromSeeds(name string, cipherSeed, signingSeed []byte) *KeyPairInfo {
	return &KeyPairInfo{
		Name:        name,
		CipherSeed:  bytes.Clone(cipherSeed),
		SigningSeed: bytes.Clone(signingSeed),
	}
}

func (kpi *KeyPairInfo) Clone() *KeyPairInfo {
	return &KeyPairInfo{
		Name:        kpi.Name,
		CipherSeed:  bytes.Clone(kpi.CipherSeed),
		SigningSeed: bytes.Clone(kpi.SigningSeed),
	}
}

func (kpi *KeyPairInfo) PrivateKeys() (cipher, signing []byte, err error) {
	if len(kpi.CipherSeed) == 0 {
		return nil, nil, errors.New("cipher seed value is empty")
	}

	cipherKP, err := nkeys.FromSeed([]byte(kpi.CipherSeed))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to derive cipher keypair: %w", err)
	}
	defer cipherKP.Wipe()

	privateKeyCipher, err := cipherKP.PrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to derive cipher private key: %w", err)
	}

	if len(kpi.SigningSeed) == 0 {
		return nil, nil, errors.New("signing seed value is empty")
	}

	signingKP, err := nkeys.FromSeed([]byte(kpi.SigningSeed))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to derive cipher keypair: %w", err)
	}
	defer signingKP.Wipe()

	privateKeySigning, err := signingKP.PrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to derive cipher private key: %w", err)
	}

	return privateKeyCipher, privateKeySigning, nil
}

func (kpi *KeyPairInfo) PublicKeys() (cipher, signing string, err error) {
	if len(kpi.CipherSeed) == 0 {
		return "", "", errors.New("cipher seed value is empty")
	}

	cipherKP, err := nkeys.FromSeed([]byte(kpi.CipherSeed))
	if err != nil {
		return "", "", fmt.Errorf("unable to derive cipher keypair: %w", err)
	}
	defer cipherKP.Wipe()

	cipherPublicKey, err := cipherKP.PublicKey()
	if err != nil {
		return "", "", fmt.Errorf("unable to derive cipher public key: %w", err)
	}

	if len(kpi.SigningSeed) == 0 {
		return "", "", errors.New("signing seed value is empty")
	}

	signingKP, err := nkeys.FromSeed([]byte(kpi.SigningSeed))
	if err != nil {
		return "", "", fmt.Errorf("unable to derive signing keypair: %w", err)
	}
	defer signingKP.Wipe()

	signingPublicKey, err := signingKP.PublicKey()
	if err != nil {
		return "", "", fmt.Errorf("unable to derive signing public key: %w", err)
	}

	return cipherPublicKey, signingPublicKey, nil
}

func (kpi *KeyPairInfo) GetCipherKeyPair() (nkeys.KeyPair, error) {
	return nkeys.FromCurveSeed(kpi.CipherSeed)
}

func (kpi *KeyPairInfo) GetSigningKeyPair() (nkeys.KeyPair, error) {
	return nkeys.FromSeed(kpi.SigningSeed)
}

func (kpi *KeyPairInfo) Print(headerText string, showAll bool) error {
	cipherKP, err := nkeys.FromSeed([]byte(kpi.CipherSeed))
	if err != nil {
		return fmt.Errorf("unable to retrieve cipher keypair from seed: %w", err)
	}
	defer cipherKP.Wipe()

	cipherPublicKey, err := cipherKP.PublicKey()
	if err != nil {
		return fmt.Errorf("unable to extract cipher public key from keypair: %w", err)
	}

	var cipherPrivateKey []byte
	if showAll {
		cipherPrivateKey, err = cipherKP.PrivateKey()
		if err != nil {
			return fmt.Errorf("unable to extract cipher private key from keypair: %w", err)
		}
	}
	defer Wipe(cipherPrivateKey)

	signingKP, err := nkeys.FromSeed([]byte(kpi.SigningSeed))
	if err != nil {
		return fmt.Errorf("unable to retrieve signing keypair from seed: %w", err)
	}
	defer signingKP.Wipe()

	signingPublicKey, err := signingKP.PublicKey()
	if err != nil {
		return fmt.Errorf("unable to extract signing public key from keypair: %w", err)
	}

	var signingPrivateKey []byte
	if showAll {
		signingPrivateKey, err = signingKP.PrivateKey()
		if err != nil {
			return fmt.Errorf("unable to extract signing private key from keypair: %w", err)
		}
	}
	defer Wipe(signingPrivateKey)

	if headerText != "" {
		fmt.Println(headerText)
		fmt.Println("=============================================")
	}

	fmt.Printf("Name: %s\n", kpi.Name)
	fmt.Println("========================================================")

	if showAll {
		fmt.Println("    Cipher Key")
		fmt.Println("    ---------------------------------------------------------")
		fmt.Printf("    KP Seed     : %s\n", string(kpi.CipherSeed))
		fmt.Printf("    Private Key : %s\n", string(cipherPrivateKey))
		fmt.Printf("    Public Key  : %s\n", cipherPublicKey)
		fmt.Println("")
		fmt.Println("    Signing Key")
		fmt.Println("    ---------------------------------------------------------")
		fmt.Printf("    KP Seed     : %s\n", string(kpi.SigningSeed))
		fmt.Printf("    Private Key : %s\n", string(signingPrivateKey))
		fmt.Printf("    Public Key  : %s\n", signingPublicKey)

		return nil

	}

	fmt.Printf("Cipher Public Key   : %s\n", cipherPublicKey)
	fmt.Printf("Signing Public Key  : %s\n", signingPublicKey)

	return nil
}

func (kpi *KeyPairInfo) Wipe() {
	if len(kpi.CipherSeed) != 0 {
		_, _ = io.ReadFull(rand.Reader, kpi.CipherSeed[:])
	}

	if len(kpi.SigningSeed) != 0 {
		_, _ = io.ReadFull(rand.Reader, kpi.SigningSeed[:])
	}
}

func (kpi *KeyPairInfo) SignRandom() ([]byte, error) {
	salt, err := helpers.GetRandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve random bytes for random sig salt: %w", err)
	}

	signedBytes, err := kpi.Sign(salt)
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

func (kpi *KeyPairInfo) Sign(input []byte) ([]byte, error) {
	signingKP, err := nkeys.FromSeed(kpi.SigningSeed)
	if err != nil {
		return nil, fmt.Errorf("failed extracting signing keypair: %w", err)
	}

	signedBytes, err := signingKP.Sign(input)
	if err != nil {
		return nil, fmt.Errorf("error signing data: %w", err)
	}

	return signedBytes, nil
}
