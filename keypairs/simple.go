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
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	"github.com/thoughtrealm/bumblebee/cipher"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"sort"
	"strings"
	"sync"
)

// SimpleKeyPairStore will be the storage file for the local account key pairs created in this profile
type SimpleKeyPairStore struct {
	// syncKeyPairs is used for concurrency safety with the map data
	syncKeyPairs sync.Mutex `msgpack:"-"`

	// KeyPairs stores the key data in a map for native lookups
	KeyPairs map[string]*KeyPairInfo

	// StoreKey stores the key used to read the keystore into this struct to use for future writes
	StoreKey []byte `msgpack:"-"`

	// OriginFilePath stores the filepath from a prior file load for future saves after keypair changes
	OriginFilePath string `msgpack:"-"`
}

func newSimpleKeyPairStore() *SimpleKeyPairStore {
	// SimpleKeyPairStore is the default
	return &SimpleKeyPairStore{
		KeyPairs: make(map[string]*KeyPairInfo),
	}
}

func (kps *SimpleKeyPairStore) Count() int {
	return len(kps.KeyPairs)
}

func (kps *SimpleKeyPairStore) SaveKeyPairStoreToOrigin(key []byte) error {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	return kps.saveKeyPairStoreToOrigin(key)
}

func (kps *SimpleKeyPairStore) saveKeyPairStoreToOrigin(key []byte) error {
	if kps.OriginFilePath == "" {
		return errors.New("unable to save keypair store: origin path is empty")
	}

	return kps.saveKeyPairStore(key, kps.OriginFilePath)
}

// SaveKeyPairStore writes the keypair data to the file path provided.
// If key is provided, it will take priority over the kps.KeyPairStoreKey value.
// If key is nil, kps.KeyPairStoreKey will be used, if it is set.
// If neither key nor kps.KeyPairStoreKey are set, it will be written out unencrypted.
func (kps *SimpleKeyPairStore) SaveKeyPairStore(key []byte, storeFilePath string) error {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	return kps.saveKeyPairStore(key, storeFilePath)
}

func (kps *SimpleKeyPairStore) saveKeyPairStore(key []byte, storeFilePath string) error {
	kpsBytes, err := msgpack.Marshal(kps)
	if err != nil {
		return fmt.Errorf("failed serializing keypair data: %w", err)
	}

	outputFile, err := os.Create(storeFilePath)
	if err != nil {
		return fmt.Errorf("failed creating keypair store file: %w", err)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	// If a key is passed in, we will overwrite the storekey.  The explicitly provided
	// key takes priority over the current storekey.
	if key != nil {
		kps.StoreKey = bytes.Clone(key)
	}

	if kps.StoreKey != nil {
		// an encryption key was provided, so encrypt and write out salt + encrypted data
		encryptedBytes, salt, err := cipher.EncryptBytes(kpsBytes, kps.StoreKey)
		if err != nil {
			return fmt.Errorf("failed encrypted keypair store data: %w", err)
		}

		_, err = cipherio.WriteBytesTo(salt, cipherio.LenMarkerSize8, outputFile)
		if err != nil {
			return fmt.Errorf("failed writing salt: %w", err)
		}

		_, err = outputFile.Write(encryptedBytes)
		if err != nil {
			return fmt.Errorf("failed writing encrypted keypair store stream: %w", err)
		}

		return err
	}

	// Write a zero marker to indicate not encrypted
	_, err = cipherio.WriteUint8Marker(0, outputFile)
	if err != nil {
		return fmt.Errorf("error writing 0 marker: %w", err)
	}

	n, err := outputFile.Write(kpsBytes)
	if err != nil {
		return fmt.Errorf("failed writing data to keystore file: %w", err)
	}
	if n != len(kpsBytes) {
		return fmt.Errorf(
			"failed writing data to keystore file. Wrote %d bytes, expected %d bytes",
			n,
			len(kpsBytes),
		)
	}

	return nil
}

// LoadKeyPairStoreFromFile checks the marker byte.  If it is 0, then the file is NOT encrypted
// and the key is ignored.  If it is NOT 0, then it attempts to decrypt the file data.
// The data is loaded into a keypair store.  Then the receiver kp store is updated in place.
// This is a replacement pattern.
// If a key is provided, it overwrites kps.StoreKey.  Otherwise, StoreKey is used if is set.
func (kps *SimpleKeyPairStore) LoadKeyPairStoreFromFile(key []byte, storeFilePath string) error {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	if !helpers.FileExists(storeFilePath) {
		return fmt.Errorf("file does not exist: \"%s\"", storeFilePath)
	}

	file, err := os.Open(storeFilePath)
	if err != nil {
		return fmt.Errorf("unable to open store file: %s", err)
	}

	defer func() {
		_ = file.Close()
	}()

	if key != nil {
		kps.StoreKey = key
	}

	fileBuff := bytes.NewBuffer(nil)
	_, err = fileBuff.ReadFrom(file)
	if err != nil {
		return fmt.Errorf("unable to read file: %s", err)
	}

	bytesStore := fileBuff.Bytes()
	if len(bytesStore) == 0 {
		return errors.New("no data retrieved from file")
	}

	var storeBytesToUse []byte
	saltLen := bytesStore[0]
	if saltLen != 0 {
		saltVal := bytesStore[1 : saltLen+1]
		storeBytesToUse, err = cipher.DecryptBytes(bytesStore[saltLen+1:], kps.StoreKey, saltVal)
		if err != nil {
			return fmt.Errorf("unable to decrypt store data: %s", err)
		}
	} else {
		// Need to remove the 0 value salt len marker
		storeBytesToUse = bytesStore[1:]
	}

	newKPS := newSimpleKeyPairStore()
	err = msgpack.Unmarshal(storeBytesToUse, newKPS)
	if err != nil {
		return fmt.Errorf("failed parsing file data: %s", err)
	}

	kps.KeyPairs = make(map[string]*KeyPairInfo)
	for name, kpi := range newKPS.KeyPairs {
		kps.KeyPairs[name] = kpi.Clone()
	}

	kps.OriginFilePath = storeFilePath

	return nil
}

func (kps *SimpleKeyPairStore) GetKeyPairInfo(name string) *KeyPairInfo {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	return kps.getKeyPairInfo(name)
}

func (kps *SimpleKeyPairStore) getKeyPairInfo(name string) *KeyPairInfo {
	// We do not lock the store here.  Internal callers lock the store first
	// as needed.

	// To make sure that the keypair info name is not case-sensitive, we convert it to an upper case value.
	kpi, found := kps.KeyPairs[strings.ToUpper(name)]
	if !found {
		return nil
	}

	return kpi.Clone()
}

func (kps *SimpleKeyPairStore) addKeyPairInfo(kpi *KeyPairInfo) {
	// We do not lock the store here.  Internal callers lock the store first
	// as needed.

	kps.KeyPairs[strings.ToUpper(kpi.Name)] = kpi.Clone()
}

func (kps *SimpleKeyPairStore) CreateNewKeyPair(name string) (*KeyPairInfo, error) {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	kpi := kps.getKeyPairInfo(name)
	if kpi != nil {
		return nil, fmt.Errorf("a keypair already exists by the name \"%s\"", name)
	}

	kp, err := nkeys.CreateCurveKeys()
	if err != nil {
		return nil, err
	}

	seed, err := kp.Seed()
	if err != nil {
		return nil, err
	}

	kpi = NewKeyPairInfo(name, seed)
	kps.addKeyPairInfo(kpi)
	return kpi, nil
}

func (kps *SimpleKeyPairStore) ListKeyPairs() []*KeyPairInfo {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	kpList := []*KeyPairInfo{}
	for _, kpi := range kps.KeyPairs {
		kpList = append(kpList, kpi.Clone())
	}

	return kpList
}

// Initialize does the following...
// - Clears the store
// - Creates and stores a new keypair with the name "default"
func (kps *SimpleKeyPairStore) Initialize() error {
	return errors.New("not implemented")
}

func (kps *SimpleKeyPairStore) RenameKeyPair(currentName, newName string) (found bool, err error) {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	kpi := kps.getKeyPairInfo(currentName)
	if kpi == nil {
		return false, nil
	}

	kpi.Name = newName

	// add the keypair using the new name for the key
	kps.KeyPairs[strings.ToUpper(kpi.Name)] = kpi

	// remove the old reference that used the old name for the key
	delete(kps.KeyPairs, strings.ToUpper(currentName))

	return true, nil
}

func (kps *SimpleKeyPairStore) RemoveKeyPair(name string) (bool, error) {
	kps.syncKeyPairs.Lock()
	defer kps.syncKeyPairs.Unlock()

	return kps.removeKeyPair(name)
}

func (kps *SimpleKeyPairStore) removeKeyPair(name string) (bool, error) {
	kpi := kps.getKeyPairInfo(name)
	if kpi == nil {
		return false, nil
	}

	delete(kps.KeyPairs, strings.ToUpper(name))
	err := kps.saveKeyPairStoreToOrigin(nil)
	if err != nil {
		return true, fmt.Errorf("unable to save keypair store to file: %w", err)
	}

	return true, nil
}

func (kps *SimpleKeyPairStore) Walk(sortInfo bool, walkFunc KeyPairStoreWalkFunc) {
	if walkFunc == nil {
		panic("walkFunc is nil")
	}

	if len(kps.KeyPairs) == 0 {
		return
	}

	if sortInfo == false {
		for _, kpi := range kps.KeyPairs {
			walkFunc(kpi)
		}
		return
	}

	// we are sorting so need to load slice with keypair info
	kplist := []*KeyPairInfo{}
	for _, kpi := range kps.KeyPairs {
		kplist = append(kplist, kpi)
	}

	sort.Slice(kplist, func(i, j int) bool {
		if helpers.CompareStrings(kplist[i].Name, kplist[j].Name) == -1 {
			return true
		}
		return false
	})

	for _, kpi := range kplist {
		walkFunc(kpi)
	}
}

func (kps *SimpleKeyPairStore) SetPassword(newPassword []byte) {
	kps.StoreKey = newPassword
}

func (kps *SimpleKeyPairStore) WipeData() {
	defer func() {
		// since this is called in possibly unstable scenarios, like during failed shutdown scenarios,
		// or just during delayed unrolling of runtime teardown,
		// let's assume that this process could always induce panics and suppress accordingly
		if r := recover(); r != nil {
			logger.Debugf("Panic in SimpleKeyPairStore WipeData(): %s", r)
		}
	}()

	for _, kpi := range kps.KeyPairs {
		kpi.Wipe()
	}
}
