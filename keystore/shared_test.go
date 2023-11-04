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

	newStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			IsDefault: true,
			KeyType:   security.KeyTypeSeed,
			Name:      "root",
			KeyData:   []byte("SXAIBZV5ONCL446HGD2OTR3NVMFY2XKXZVXQX7ARKGBJMM32WG2G2BHXYU"),
		})

	_ = newStore.AddEntity("bob", &security.KeyInfo{
		IsDefault: true,
		KeyType:   security.KeyTypePublic,
		Name:      "mykey",
		KeyData:   []byte("iso public"),
	})

	return
}

func buildTestStoreMultiEntity(countOfEntities int) (newStore *SimpleKeyStore) {
	newStore = newSimpleKeyStore("local", "local", true)

	newStore.SetServerInfo(
		"name",
		"localhost",
		"16222",
		&security.KeyInfo{
			IsDefault: true,
			KeyType:   security.KeyTypeSeed,
			Name:      "root",
			KeyData:   []byte("SXAIBZV5ONCL446HGD2OTR3NVMFY2XKXZVXQX7ARKGBJMM32WG2G2BHXYU"),
		})

	for count := 0; count < countOfEntities; count++ {
		_ = newStore.AddEntity(fmt.Sprintf("bob-%d", count), &security.KeyInfo{
			IsDefault: true,
			KeyType:   security.KeyTypePublic,
			Name:      fmt.Sprintf("mykey-%d", count),
			KeyData:   []byte(fmt.Sprintf("iso public-%d", count)),
		})
	}

	return
}
