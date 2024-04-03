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

package bootstrap

import (
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/logger"
)

type BootstrapLoader struct {
	ActiveProfile   *helpers.Profile
	KeyPairStoreKey []byte
}

func Run(loadKeystore, loadKeypairStore bool) error {
	if helpers.CmdHelpers.UseProfile != "" {
		helpers.GlobalUseProfile = helpers.CmdHelpers.UseProfile
		logger.Debugfln("Using profile \"%s\"", helpers.GlobalUseProfile)
	}

	loader := &BootstrapLoader{}

	if err := loader.loadConfig(loadKeypairStore); err != nil {
		return err
	}

	// we load the keypair store first, because it is needed to decrypt the key store
	if loadKeypairStore {
		err := loader.checkKeyPairStoreKey()
		if err != nil {
			return fmt.Errorf("unable to check key state for keypair store: %w", err)
		}

		if err = loader.loadGlobalKeypairStore(); err != nil {
			return fmt.Errorf("unable to load keypair store: %w", err)
		}
	}

	if loadKeystore {
		if err := loader.loadGlobalKeyStore(); err != nil {
			return fmt.Errorf("unable to load keystore: %w", err)
		}
	}

	return nil
}

func (bsl *BootstrapLoader) loadConfig(loadActiveProfile bool) error {
	helpers.GlobalConfig = helpers.NewConfigHelper()
	err := helpers.GlobalConfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("unable to load config during startup: %w", err)
	}

	if loadActiveProfile {
		bsl.ActiveProfile = helpers.GlobalConfig.GetCurrentProfile()
		if bsl.ActiveProfile == nil {
			return errors.New("current profile was not located")
		}
	}

	return nil
}

func (bsl *BootstrapLoader) checkKeyPairStoreKey() error {
	var err error
	if bsl.ActiveProfile.KeyPairStoreEncrypted {
		bsl.KeyPairStoreKey, err = helpers.AcquireKey(bsl.ActiveProfile.Name)
		if err != nil {
			return fmt.Errorf("unable to acquire key for decrypting the keystore: %w", err)
		}

		if len(bsl.KeyPairStoreKey) == 0 {
			return errors.New("profile indicates keypair store is encrypted, but no key was acquired")
		}
	}

	return nil
}

func (bsl *BootstrapLoader) loadGlobalKeyStore() error {
	keystorePath := bsl.ActiveProfile.KeyStorePath

	var err error
	keystore.GlobalKeyStore, err = keystore.NewFromFile(bsl.KeyPairStoreKey, keystorePath)
	if err != nil {
		return fmt.Errorf("unable to load keystore file: %w", err)
	}

	return nil
}

func (bsl *BootstrapLoader) loadGlobalKeypairStore() error {
	keypairStorePath := bsl.ActiveProfile.KeyPairStorePath

	// for now, we will assume that we are using the same key for both stores.
	// so we won't check for it again, since it should have been gathered
	// when loading the keystore
	var err error
	keypairs.GlobalKeyPairStore, err = keypairs.NewKeypairStoreFromFile(bsl.KeyPairStoreKey, keypairStorePath)
	if err != nil {
		return fmt.Errorf("unable to load keypair store file: %s", err)
	}

	return nil
}
