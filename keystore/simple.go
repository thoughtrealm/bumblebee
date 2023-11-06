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
	"github.com/nats-io/nkeys"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/vmihailenco/msgpack/v5"
	"sort"
	"sync"
)

type SimpleKeyStore struct {
	// Todo: make sure this SyncStore mutex is being used where needed
	Details        *StoreDetails
	Server         *ServerInfo
	Entities       EntityCollection
	SyncStore      sync.RWMutex `msgpack:"-"`
	SourceFilePath string       `msgpack:"-"`
}

// New returns a new SimpleKeyStore with no entities and name, owner and isLocal set accordingly
func newSimpleKeyStore(name, owner string, isLocal bool) *SimpleKeyStore {
	return &SimpleKeyStore{
		Details:  &StoreDetails{Name: name, Owner: owner, IsLocal: isLocal, IsDirty: true},
		Server:   &ServerInfo{},
		Entities: map[string]*security.Entity{},
	}
}

// NewFromMemory returns a new keystore from the byte sequence.
// The byte sequence would be bytes read from a keystore file, etc.
func newSimpleKeyStoreFromMemory(bytesStore []byte) (newKeyStore *SimpleKeyStore, err error) {
	newKeyStore = &SimpleKeyStore{}
	err = msgpack.Unmarshal(bytesStore, newKeyStore)
	if err != nil {
		return nil, err
	}
	return newKeyStore, nil
}

// NewFromFile returns a new keystore read from a file.
// The key sequence is used if the file is encrypted.  If it is not encrypted, the key is ignored.
func newSimpleKeyStoreFromFile(filePath string) (newKeyStore *SimpleKeyStore, err error) {
	newKeyStore = &SimpleKeyStore{}
	err = newKeyStore.ReadFromFile(filePath)
	if err != nil {
		return nil, err
	}
	newKeyStore.SourceFilePath = filePath
	return newKeyStore, nil
}

func (sks *SimpleKeyStore) NewKeyPair() (kp nkeys.KeyPair, err error) {
	return nkeys.CreateCurveKeys()
}

func (sks *SimpleKeyStore) AddKey(name string, publicKey []byte) error {
	newKeyInfo, err := security.NewKeyInfo(false, security.KeyTypePublic, name, publicKey)
	if err != nil {
		return fmt.Errorf("unable to make new keyInfo: %w", err)
	}

	err = sks.AddEntity(name, newKeyInfo)
	if err != nil {
		return fmt.Errorf("unable to add new Entity: %w", err)
	}

	// providing empty values tells WriteToFile to use the corresponding fields of sks
	err = sks.WriteToFile("")
	if err != nil {
		return fmt.Errorf("unable to write updated keystore to file: %w", err)
	}

	return nil
}

func (sks *SimpleKeyStore) GetKey(name string) *security.Entity {
	entity := sks.GetEntity(name)
	return entity
}

func (sks *SimpleKeyStore) Rename(newName string) error {
	if newName == "" {
		return errors.New("value of supplied target name is empty")
	}

	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	sks.Details.Name = newName

	// Todo: In the future... we will want a flag that indicates to rename the storage file as per the new store name
	return nil
}

// Load takes a source key store and replaces the data in the reference keystore with the
// source data.  This is a replacement pattern that just updates the referenced keystore.
func (sks *SimpleKeyStore) Load(sourceKeyStore *SimpleKeyStore) {
	sks.Details = sourceKeyStore.Details.Clone()
	sks.Server = sourceKeyStore.Server.Clone()
	sks.Entities = sourceKeyStore.Entities.Clone()
	sks.SourceFilePath = sourceKeyStore.SourceFilePath
}

func (sks *SimpleKeyStore) Count() int {
	return len(sks.Entities)
}

func (sks *SimpleKeyStore) WalkCount(nameMatchFilter string, walkFilterFunc KeyStoreWalkFilterFunc) (count int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in KeyStore.Walk: %s", r)
		}
	}()

	// If no entities, then just return
	if len(sks.Entities) == 0 {
		return 0, nil
	}

	for _, entity := range sks.Entities {
		shouldInclude := true
		if nameMatchFilter != "" {
			shouldInclude = helpers.MatchesFilter(entity.Name, nameMatchFilter)
			if !shouldInclude {
				// If the name match excludes the entity, then we can escape now,
				// even if there is a filter func.  The name filter will take priority.
				continue
			}
		}

		if walkFilterFunc != nil {
			shouldInclude = walkFilterFunc(entity)
			if !shouldInclude {
				continue
			}
		}

		count += 1
	}

	return count, nil
}

func (sks *SimpleKeyStore) Walk(walkInfo *WalkInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in KeyStore.Walk: %s", r)
		}
	}()

	// If no entities, then just return
	if len(sks.Entities) == 0 {
		return
	}

	// At a bare minimum, the caller must provide a WalkFunc
	if walkInfo.WalkFunc == nil {
		return errors.New("no WalkFunc provided")
	}

	// If they have not requested anyt filtering or sorting logic, then we can
	// short circuit and just iterate over the map itself and escape
	if walkInfo.NameMatchFilter == "" && walkInfo.WalkFilterFunc == nil && walkInfo.SortResults == false {
		for _, entity := range sks.Entities {
			walkInfo.WalkFunc(entity)
		}

		return nil
	}

	// We have filtering and/or sorting logic, so we must move the map data into a slice
	var entities []*security.Entity
	if walkInfo.WalkFilterFunc == nil {
		// no filtering logic provided, so make entities the size of the map
		entities = make([]*security.Entity, 0, len(sks.Entities))
	} else {
		// otherwise, since filtering is requested, we don't know what the final size
		// will be, just >= 0.  So, set size to 0 and grow as needed.
		entities = make([]*security.Entity, 0)
	}

	for _, entity := range sks.Entities {
		shouldInclude := true
		if walkInfo.NameMatchFilter != "" {
			shouldInclude = helpers.MatchesFilter(entity.Name, walkInfo.NameMatchFilter)
			if !shouldInclude {
				// If the name match excludes the entity, then we can escape now,
				// even if there is a filter func.  The name filter will take priority.
				continue
			}
		}

		if walkInfo.WalkFilterFunc != nil {
			shouldInclude = walkInfo.WalkFilterFunc(entity)
			if !shouldInclude {
				continue
			}
		}

		entities = append(entities, entity)
	}

	if len(entities) == 0 {
		return nil
	}

	// if a sort func is provided, then sort the slice values
	if walkInfo.SortResults == true {
		sort.Slice(entities, func(i, j int) bool {
			if helpers.CompareStrings(entities[i].Name, entities[j].Name) == -1 {
				return true
			}

			return false
		})
	}

	// now, we can iterate over the slice
	for _, entity := range entities {
		walkInfo.WalkFunc(entity)
	}

	return nil
}

func (sks *SimpleKeyStore) WipeData() {
	defer func() {
		// since this is called in possibly unstable scenarios, like during failed shutdown scenarios,
		// or just during delayed unrolling of runtime teardown,
		// let's assume that this process could always induce panics and suppress accordingly
		if r := recover(); r != nil {
			logger.Debugf("Panic in SimpleKeyStore WipeData(): %s", r)
		}
	}()

	if sks == nil {
		return
	}

	for _, entity := range sks.Entities {
		if entity.Key != nil {
			entity.Key.Wipe()
		}
	}
}
