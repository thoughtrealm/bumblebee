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
	"github.com/thoughtrealm/bumblebee/security"
)

func buildTestStore() (newStore *SimpleKeyStore) {
	newStore = newSimpleKeyStore("local", "local", true)

	storeKPI, _ := security.NewKeyPairInfoWithSeeds("store")
	storeCipherPubKey, storeSigningPubKey, _ := storeKPI.PublicKeys()

	entityKPI, _ := security.NewKeyPairInfoWithSeeds("entity")
	entityCipherPubKey, entitySigningPubKey, _ := entityKPI.PublicKeys()
	entityKI, _ := security.NewKeyInfo(
		"entityKI",
		entityCipherPubKey,
		entitySigningPubKey,
	)

	newStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			Name:          "root",
			CipherPubKey:  storeCipherPubKey,
			SigningPubKey: storeSigningPubKey,
		},
	)

	_ = newStore.AddEntity("bob", entityKI)

	return
}

func buildTestStoreMultiEntity(countOfEntities int) (newStore *SimpleKeyStore) {
	newStore = newSimpleKeyStore("local", "local", true)

	storeKPI, _ := security.NewKeyPairInfoWithSeeds("store")
	storeCipherPubKey, storeSigningPubKey, _ := storeKPI.PublicKeys()

	newStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			Name:          "root",
			CipherPubKey:  storeCipherPubKey,
			SigningPubKey: storeSigningPubKey,
		},
	)

	for count := 0; count < countOfEntities; count++ {
		entityKPI, _ := security.NewKeyPairInfoWithSeeds(fmt.Sprintf("mykey-%d", count))
		entityCipherPubKey, entitySigningPubKey, _ := entityKPI.PublicKeys()
		entityKI, _ := security.NewKeyInfo(
			fmt.Sprintf("mykey-%d", count),
			entityCipherPubKey,
			entitySigningPubKey,
		)

		_ = newStore.AddEntity(fmt.Sprintf("bob-%d", count), entityKI)
	}

	return
}
