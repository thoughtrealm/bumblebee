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
	"bytes"
	"errors"
	"fmt"
	"github.com/nats-io/nkeys"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/vmihailenco/msgpack/v5"
	"os"
)

// ReadFromFile will call the cipher IO to decrypt the combined bundle stream from the keystore file
func (sks *SimpleKeyStore) ReadFromFile(filePath string) error {
	// we need to get the read kpi from the global keypair store
	kpiRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreReads)
	if kpiRead == nil {
		return errors.New("keypair for keystore reads was not found in the global keypair store")
	}

	// Convert the read kpi to a receiver kpi for the cipher reader
	newKPIReceiver, err := security.NewKeyInfo(false, security.KeyTypeSeed, "reader", []byte(kpiRead.Seed))
	if err != nil {
		return fmt.Errorf("unable to transform read keypair info to receiver info: %w", err)
	}

	// Now get the write KPI from the keypair store
	kpiWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiWrite == nil {
		return errors.New("keypair for keystore writes not found in the global keypair store")
	}

	// now, extract the writer KP from the seed
	kpWrite, err := nkeys.FromSeed([]byte(kpiWrite.Seed))
	if err != nil {
		return fmt.Errorf("unable to transform write keypair info to keypair: %w", err)
	}

	// now get the writer public key from the KP for writes
	pubKey, err := kpWrite.PublicKey()
	if err != nil {
		return fmt.Errorf("unable to extract public key from write keypair: %w", err)
	}

	// Now build the sender KPI for the cipher reader
	newKPISender, err := security.NewKeyInfo(false, security.KeyTypePublic, "writer", []byte(pubKey))

	cfr, err := cipherio.NewCipherFileReader(newKPIReceiver, newKPISender)
	if err != nil {
		return fmt.Errorf("unable to create instance of cipher reader: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open keystore file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	storeBytes, err := cfr.ReadCombinedFileToBytes(filePath)
	if err != nil {
		return fmt.Errorf("unable to read keystore data from file: %w", err)
	}

	sourceKeyStore := &SimpleKeyStore{}
	err = msgpack.Unmarshal(storeBytes, sourceKeyStore)
	if err != nil {
		return fmt.Errorf("failed interpreting keystore byte sequence: %w", err)
	}

	// Update the current keystore with the memKeyStore data
	sks.Load(sourceKeyStore)
	return nil
}

func (sks *SimpleKeyStore) initializeCipherWriter() (*cipherio.CipherWriter, error) {
	// we need to get the read kpi from the global keypair store
	kpiRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreReads)
	if kpiRead == nil {
		return nil, errors.New("keypair for keystore reads was not found in the global keypair store")
	}

	// now, extract the reader KP from the seed
	kpRead, err := nkeys.FromSeed([]byte(kpiRead.Seed))
	if err != nil {
		return nil, fmt.Errorf("unable to transform read keypair info to keypair: %w", err)
	}

	// now get the reader public key from the KP for reads
	pubKey, err := kpRead.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("unable to extract public key from read keypair: %w", err)
	}

	// Convert the read kpi to a receiver kpi for the cipher reader
	newKPIReceiver, err := security.NewKeyInfo(false, security.KeyTypePublic, "reader", []byte(pubKey))
	if err != nil {
		return nil, fmt.Errorf("unable to transform read keypair info to receiver info: %w", err)
	}

	// Now get the write KPI from the keypair store
	kpiWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiWrite == nil {
		return nil, errors.New("keypair for keystore writes not found in the global keypair store")
	}

	// Now build the sender KPI for the cipher reader
	newKPISender, err := security.NewKeyInfo(false, security.KeyTypeSeed, "writer", []byte(kpiWrite.Seed))

	cfw, err := cipherio.NewCipherWriter(newKPIReceiver, newKPISender)
	if err != nil {
		return nil, fmt.Errorf("unable to create instance of cipher writer: %w", err)
	}

	return cfw, nil
}

// WriteToFile writes the keystore data to a file.
//   - filePath should include the entire path ref including both path and name
//   - if filePath is empty, it will check the SourceFilePath member of SimpleKeyStore
func (sks *SimpleKeyStore) WriteToFile(filePath string) error {
	if filePath == "" && sks.SourceFilePath == "" {
		return errors.New("no target file path provided and no prior filepath available")
	}

	var useFilePath string
	if filePath == "" {
		useFilePath = sks.SourceFilePath
	} else {
		useFilePath = filePath
	}

	cfw, err := sks.initializeCipherWriter()
	if err != nil {
		return fmt.Errorf("unable to initialize a cipher writer: %w", err)
	}

	bytesStore, err := msgpack.Marshal(sks)
	if err != nil {
		return fmt.Errorf("unable to serialize keystore data: %w", err)
	}

	outputBuffer := bytes.NewBuffer(bytesStore)
	_, err = cfw.WriteToCombinedFileFromReader(useFilePath, outputBuffer)
	if err != nil {
		return fmt.Errorf("unable to write keystore to file: %w", err)
	}

	sks.Details.IsDirty = false
	return nil
}

func (sks *SimpleKeyStore) WriteToMemory() (bytesStore []byte, err error) {
	return msgpack.Marshal(sks)
}

/*
// LoadFromMemory updates the current sks using data from the streamed key store.
//   - It should check to see if the stream is encrypted if checkForEncryption is set.
//   - If it is NOT encrypted, then the leading byte will be a 0 marker.
//   - If it is encrypted, then the first byte will be anything other than 0.
//   - If it is encrypted, then the key should be in sks.KeyPairStoreKey.
func (sks *SimpleKeyStore) LoadFromMemory(bytesStore []byte, checkForEncryption bool) error {
	if len(bytesStore) == 0 {
		return errors.New("no data provided: bytesStore has length 0")
	}

	memKeyStore := &SimpleKeyStore{}

	var storeBytesToUse []byte
	var err error

	saltLen := bytesStore[0]
	if checkForEncryption && saltLen != 0 {
		saltVal := bytesStore[1 : saltLen+1]
		storeBytesToUse, err = cipher.DecryptBytes(bytesStore[saltLen+1:], sks.KeyPairStoreKey, saltVal)
		if err != nil {
			return fmt.Errorf("unable to decrypt store data: %s", err)
		}

	} else {
		// Need to remove the 0 value salt len marker
		storeBytesToUse = bytesStore[1:]
	}

	err = msgpack.Unmarshal(storeBytesToUse, memKeyStore)
	if err != nil {
		return fmt.Errorf("failed interpreting byte sequence: %s", err)
	}

	// Update the current keystore with the memKeyStore data
	sks.Load(memKeyStore)
	return nil
}
*/
