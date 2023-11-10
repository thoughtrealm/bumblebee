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

	storeKPI, _ := security.NewKeyPairInfoWithSeeds("store")
	storeCipherPubKey, storeSigningPubKey, _ := storeKPI.PublicKeys()

	s.testStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			Name:          "root",
			CipherPubKey:  storeCipherPubKey,
			SigningPubKey: storeSigningPubKey,
		},
	)
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

func (s *KeyStoreTestSuite) TestSimpleKeyStore_DeriveServerPublicKey() {
	server := s.testStore.GetServerInfo()
	if !s.NotNil(server) {
		return
	}

	fmt.Printf("CipherPublicKey  : %s\n", server.Key.CipherPubKey)
	fmt.Printf("SigningPublicKey : %s\n", server.Key.SigningPubKey)
}
