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
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/vmihailenco/msgpack/v5"
	"os"
)

// ReadFromFile will call the cipher IO to decrypt the combined bundle stream from the keystore file
func (sks *SimpleKeyStore) ReadFromFile(filePath string) error {
	// We need to get the read kpi from the global keypair store, which we store in the cipher reader,
	// so we dont wipe it here.  It's a clone, so it will be wiped later with cipher file reader.
	kpiRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreReads)
	if kpiRead == nil {
		return errors.New("keypair info for keystore reads was not found in the global keypair store")
	}

	// Now get the writer KPI from the keypair store
	kpiWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiWrite == nil {
		return errors.New("keypair info for keystore writes not found in the global keypair store")
	}
	defer kpiWrite.Wipe()

	// now, extract the writer KP from the seed
	writerCipherPubKey, writerSigningPubKey, err := kpiWrite.PublicKeys()
	if err != nil {
		return fmt.Errorf("unable to retrieve pubkeys from writer kpi: %w", err)
	}

	// Now build the sender KPI for the cipher reader
	newKISender, err := security.NewKeyInfo("writer", writerCipherPubKey, writerSigningPubKey)
	if err != nil {
		return fmt.Errorf("unable to construct sender keyinfo from writer pubkeys: %w", err)
	}

	cfr, err := cipherio.NewCipherFileReader(kpiRead, newKISender)
	if err != nil {
		return fmt.Errorf("unable to create instance of cipher reader: %w", err)
	}
	defer cfr.Wipe()

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
	defer security.Wipe(storeBytes)

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
	defer kpiRead.Wipe()

	cipherPubKey, signingPubKey, err := kpiRead.PublicKeys()
	if err != nil {
		return nil, fmt.Errorf("unable to extract pub keys from reader kpi: %w", err)
	}

	// Convert the read kpi to a receiver ki for the cipher reader
	newKIReceiver, err := security.NewKeyInfo("reader", cipherPubKey, signingPubKey)
	if err != nil {
		return nil, fmt.Errorf("unable to transform read pubkeys to receiver keyinfo: %w", err)
	}

	// Now get the write KPI from the keypair store
	kpiWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiWrite == nil {
		return nil, errors.New("keypair for keystore writes not found in the global keypair store")
	}
	defer kpiWrite.Wipe()

	cfw, err := cipherio.NewCipherWriter(newKIReceiver, kpiWrite)
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
	defer security.Wipe(bytesStore)

	readBuffer := bytes.NewReader(bytesStore)
	_, err = cfw.WriteToCombinedFileFromReader(useFilePath, readBuffer)
	if err != nil {
		return fmt.Errorf("unable to write keystore to file: %w", err)
	}

	sks.Details.IsDirty = false
	return nil
}

func (sks *SimpleKeyStore) WriteToMemory() (bytesStore []byte, err error) {
	return msgpack.Marshal(sks)
}
