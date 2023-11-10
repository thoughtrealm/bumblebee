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
	"fmt"
	"github.com/thoughtrealm/bumblebee/security"
)

var GlobalKeyPairStore KeyPairStore

type KeyPairStoreWalkFunc func(kpi *security.KeyPairInfo)

type KeyPairStore interface {
	Count() int
	CreateNewKeyPair(name string) (*security.KeyPairInfo, error)
	GetKeyPairInfo(name string) *security.KeyPairInfo
	LoadKeyPairStoreFromFile(key []byte, filePath string) error
	RemoveKeyPair(name string) (bool, error)
	RenameKeyPair(currentName, newName string) (found bool, err error)
	SaveKeyPairStore(key []byte, storeFilePath string) error
	SaveKeyPairStoreToOrigin(key []byte) error
	SetPassword(newPassword []byte)
	Walk(sort bool, walkFunc KeyPairStoreWalkFunc)
	WipeData()
}

func NewKeypairStore() KeyPairStore {
	// SimpleKeyPairStore is the default
	return newSimpleKeyPairStore()
}

func NewKeypairStoreWithKeypair(name string) (KeyPairStore, *security.KeyPairInfo, error) {
	newKPS := NewKeypairStore()
	kpName := name
	if name == "" {
		kpName = "default"
	}

	kpi, err := newKPS.CreateNewKeyPair(kpName)
	if err != nil {
		return nil, nil, err
	}

	return newKPS, kpi, nil
}

func NewKeypairStoreFromFile(key []byte, filePath string) (KeyPairStore, error) {
	kps := NewKeypairStore()
	err := kps.LoadKeyPairStoreFromFile(key, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to load keypair store file: %w", err)
	}

	return kps, nil
}

func WipeGlobalKeyPairsIfValid() {
	if GlobalKeyPairStore != nil {
		GlobalKeyPairStore.WipeData()
		GlobalKeyPairStore = nil
	}
}
