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
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/security"
	"strings"
)

type EntityCollection map[string]*security.Entity

// NewEntities returns an empty map of type map[string]*Entity
func NewEntities() EntityCollection {
	return map[string]*security.Entity{}
}

func (ec EntityCollection) Clone() EntityCollection {
	ecOutput := NewEntities()

	for id, entity := range ec {
		ecOutput[id] = entity.Clone()
	}

	return ecOutput
}

func (sks *SimpleKeyStore) AddEntity(name string, key *security.KeyInfo) error {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	if sks.getEntity(name) != nil {
		return fmt.Errorf("an entity already exists with name %s", name)
	}

	sks.Entities[strings.ToUpper(name)] = &security.Entity{
		Name:       name,
		PublicKeys: key.Clone(),
	}

	sks.Details.IsDirty = true
	return nil
}

func (sks *SimpleKeyStore) RenameEntity(oldName, newName string) (bool, error) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	entity := sks.getEntity(oldName)
	if entity == nil {
		return false, fmt.Errorf("entity not found with name \"%s\"", oldName)
	}

	entity.Name = newName
	entity.PublicKeys.Name = newName

	// Update map key entry with new name
	sks.Entities[strings.ToUpper(newName)] = entity

	// Remove prior entity with old name
	delete(sks.Entities, strings.ToUpper(oldName))

	sks.Details.IsDirty = true
	return true, nil
}

func (sks *SimpleKeyStore) UpdatePublicKeys(name, cipherPublicKey, signingPublicKey string) (found bool, err error) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	entity := sks.getEntity(name)
	if entity == nil {
		return false, fmt.Errorf("entity not found with name \"%s\"", name)
	}

	if cipherPublicKey == "" {
		return true, errors.New("provided cipherPublicKey is empty")
	}

	if signingPublicKey == "" {
		return true, errors.New("provided signingPublicKey is empty")
	}

	entity.PublicKeys.CipherPubKey = cipherPublicKey
	entity.PublicKeys.SigningPubKey = signingPublicKey
	sks.Details.IsDirty = true
	return true, nil
}

func (sks *SimpleKeyStore) UpdateCipherPublicKey(name string, cipherPublicKey string) (found bool, err error) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	entity := sks.getEntity(name)
	if entity == nil {
		return false, fmt.Errorf("entity not found with name \"%s\"", name)
	}

	if cipherPublicKey == "" {
		return true, errors.New("provided cipherPublicKey is empty")
	}

	entity.PublicKeys.CipherPubKey = cipherPublicKey
	sks.Details.IsDirty = true
	return true, nil
}

func (sks *SimpleKeyStore) UpdateSigningPublicKey(name string, signingPublicKey string) (found bool, err error) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	entity := sks.getEntity(name)
	if entity == nil {
		return false, fmt.Errorf("entity not found with name \"%s\"", name)
	}

	if signingPublicKey == "" {
		return true, errors.New("provided signingPublicKey is empty")
	}

	entity.PublicKeys.SigningPubKey = signingPublicKey
	sks.Details.IsDirty = true
	return true, nil
}

func (sks *SimpleKeyStore) GetEntity(name string) (outEntity *security.Entity) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	actualEntity := sks.getEntity(name)
	if actualEntity == nil {
		return actualEntity
	}

	outEntity = &security.Entity{
		Name:       actualEntity.Name,
		PublicKeys: actualEntity.PublicKeys.Clone(),
	}

	return outEntity
}

func (sks *SimpleKeyStore) getEntity(name string) (outEntity *security.Entity) {
	if name == "" {
		return nil
	}

	return sks.Entities[strings.ToUpper(name)]
}
func (sks *SimpleKeyStore) RemoveEntity(name string) (found bool, err error) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	return sks.removeEntity(name)
}

func (sks *SimpleKeyStore) removeEntity(name string) (found bool, err error) {
	if name == "" {
		return false, errors.New("provided entity name is empty")
	}

	if _, found = sks.Entities[strings.ToUpper(name)]; found == false {
		return false, nil
	}

	delete(sks.Entities, strings.ToUpper(name))
	sks.Details.IsDirty = true
	return true, nil
}
