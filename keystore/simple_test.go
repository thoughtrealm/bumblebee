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

package keystore

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/thoughtrealm/bumblebee/security"
	"testing"
)

type KeyStoreTestSuite struct {
	suite.Suite
	testStore *SimpleKeyStore
}

func TestKeyStoreTestSuite(t *testing.T) {
	suite.Run(t, new(KeyStoreTestSuite))
}

func (s *KeyStoreTestSuite) SetupTest() {
	s.testStore = newSimpleKeyStore("local", "local", true)
	s.testStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			IsDefault: true,
			KeyType:   security.KeyTypeSeed,
			Name:      "root",
			KeyData:   []byte("SXAIBZV5ONCL446HGD2OTR3NVMFY2XKXZVXQX7ARKGBJMM32WG2G2BHXYU"),
		})
}

func (s *KeyStoreTestSuite) TestSimpleKeyStore_GetServerByName() {
	server := s.testStore.GetServerInfo()
	s.NotNil(server)
}

func (s *KeyStoreTestSuite) TestSimpleKeyStore_NewKeyPair() {
	kp, err := s.testStore.NewKeyPair()
	if !s.Nil(err) {
		return
	}

	seed, err := kp.Seed()
	if !s.Nil(err) {
		return
	}

	fmt.Printf("key seed: %s\n", seed)
}

func (s *KeyStoreTestSuite) TestSimpleKeyStore_DerivePublicKey() {
	server := s.testStore.GetServerInfo()
	if !s.NotNil(server) {
		return
	}

	public, err := server.Key.PublicKey()
	if !s.Nil(err) {
		return
	}

	fmt.Printf("PublicKey: %s\n", string(public))
}

func (s *KeyStoreTestSuite) TestSimpleKeyStore_DerivePrivateKey() {
	server := s.testStore.GetServerInfo()
	if !s.NotNil(server) {
		return
	}

	private, err := server.Key.PrivateKey()
	if !s.Nil(err) {
		return
	}

	fmt.Printf("PrivateKey: %s\n", string(private))
}
