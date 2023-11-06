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

package keypairs

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"github.com/thoughtrealm/bumblebee/security"
	"io"
)

// KeyPairInfo defines the local environment's keypairs created for sending and receiving bundles
type KeyPairInfo struct {
	// When initialized, default is the main key.
	// More keys can be added with user-defined names.
	Name string

	// Seed is stored as NATS base32 string
	Seed []byte
}

func NewKeyPairInfo(name string, seed []byte) *KeyPairInfo {
	return &KeyPairInfo{
		Name: name,
		Seed: bytes.Clone(seed),
	}
}

func (kpi *KeyPairInfo) Clone() *KeyPairInfo {
	return &KeyPairInfo{
		Name: kpi.Name,
		Seed: bytes.Clone(kpi.Seed),
	}
}

func (kpi *KeyPairInfo) PrivateKey() ([]byte, error) {
	if len(kpi.Seed) == 0 {
		return nil, errors.New("seed value is empty")
	}

	kp, err := nkeys.FromSeed([]byte(kpi.Seed))
	if err != nil {
		return nil, fmt.Errorf("unable to derive keypair: %w", err)
	}
	defer kp.Wipe()

	privateKey, err := kp.PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to derive private key: %w", err)
	}

	return privateKey, nil
}

func (kpi *KeyPairInfo) PublicKey() ([]byte, error) {
	if len(kpi.Seed) == 0 {
		return nil, errors.New("seed value is empty")
	}

	kp, err := nkeys.FromSeed([]byte(kpi.Seed))
	if err != nil {
		return nil, fmt.Errorf("unable to derive keypair: %w", err)
	}
	defer kp.Wipe()

	publicKey, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("unable to derive public key: %w", err)
	}

	return []byte(publicKey), nil
}

func (kpi *KeyPairInfo) Print(headerText string, showAll bool) error {
	kp, err := nkeys.FromSeed([]byte(kpi.Seed))
	if err != nil {
		return fmt.Errorf("unable to retrieve keypair from seed: %w", err)
	}
	defer kp.Wipe()

	publicKey, err := kp.PublicKey()
	if err != nil {
		return fmt.Errorf("unable to extract public key from keypair: %w", err)
	}

	var privateKey []byte
	if showAll {
		privateKey, err = kp.PrivateKey()
		if err != nil {
			return fmt.Errorf("unable to extract private key from keypair: %w", err)
		}
	}
	defer security.Wipe(privateKey)

	if headerText != "" {
		fmt.Println(headerText)
		fmt.Println("=============================================")
	}

	fmt.Printf("Name        : %s\n", kpi.Name)

	if showAll {
		fmt.Printf("KP Seed     : %s\n", kpi.Seed)
		fmt.Printf("Private Key : %s\n", string(privateKey))
		fmt.Printf("Public Key  : %s\n", publicKey)

		return nil

	}

	fmt.Printf("Public Key  : %s\n", publicKey)

	return nil
}

func (kpi *KeyPairInfo) Wipe() {
	if len(kpi.Seed) != 0 {
		_, _ = io.ReadFull(rand.Reader, kpi.Seed[:])
	}
}
