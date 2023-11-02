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

package env

import (
	"errors"
	"fmt"
	"github.com/kirsle/configdir"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/keystore"
	"io/fs"
	"os"
	"path/filepath"
)

// InitializeEnvironment should...
// - Confirm a name for the default profile
// - Create the profile defs file
func InitializeEnvironment() error {
	if shouldAbort := AssertEnvironmentIsEmpty(); shouldAbort == true {
		// User decided NOT to initalize the environment, so just return
		return fmt.Errorf("user declined to delete current environment components")
	}

	// After calling AssertEnvironmentIsEmpty above, we should be able to proceed forward
	// with doing anything in the main local config environment we want to do now.  So,
	// we won't check anymore to see if it is ok to create elements there.  We will
	// assume it's a blank slate now.

	// create a default, empty profile def file
	config := &helpers.ConfigInfo{
		Profiles: []*helpers.Profile{},
	}

	fmt.Println("Saving empty profile definitions...")
	ch := helpers.NewConfigHelperFromConfig(config)
	err := ch.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable to write out profile metadata template: %w", err)
	}

	profileName, err := helpers.GetConsoleInputLine("Enter a name for the default profile (empty input for \"default\")")
	if err != nil {
		return fmt.Errorf("failed during input of profile name: %w", err)
	}

	if profileName == "" {
		profileName = "default"
	}

	return buildNewProfile(profileName)
}

func CreateNewProfile(profileName string) error {
	var err error

	// if not profile name is passed in, we will ask for it
	fmt.Println()
	if profileName == "" {
		profileName, err = helpers.GetConsoleInputLine("Enter a name for the new profile")
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed during input of profile name: %w", err)
		}
	}

	if profileName == "" {
		return fmt.Errorf("unable to create new profile: %s", "no profile name provided. Profile name is required")
	}

	return buildNewProfile(profileName)
}

// buildNewProfile does the following...
//   - Set the current profile context to the default.
//   - Create the profile path.
//   - Create the profile keypair and store it in a keypairs file, get an optional symmetric password for it.
//   - Create an empty keystore and save it, as a combined file bundle.
func buildNewProfile(profileName string) error {
	safeProfileName := helpers.GetSafeFileName(profileName)

	fmt.Println("Constructing profile path...")
	profilePath, err := helpers.BuildProfilePath(safeProfileName)
	if err != nil {
		// BuildProfilePath constructs the full error message so just return that here
		return err
	}
	fmt.Println()

	profileKeystorePath := filepath.Join(profilePath, safeProfileName+".keystore")
	profileKeypairStorePath := filepath.Join(profilePath, safeProfileName+".keypairs")

	fmt.Println(`Enter a password/key to encrypt the keypair store file for this profile. You may leave this value blank
if you do not want the keypair file encrypted.  The file is stored in your personal user config space and
you may choose to leave it unencrypted.  Adding a password is an additional layer of protection.

If you do provide a password, you will be required to provide this password each time you run bee.  
Please review the help info for how you may provide this key using environment variables and other 
mechanisms.

If you lose or forget this key, you will not be able to access the keypair file anymore, which will also result in losing any
public keys you have stored for other users.  In the case of losing the password, you will be
required to re-initialize this profile.`)

	fmt.Printf("\nEnter a password/key if you want... or leave this empty for no key: ")
	storeKey, err := helpers.GetPasswordWithConfirm("")
	fmt.Println("")
	if err != nil {
		return fmt.Errorf("failed while entering a password: %w", err)
	}

	fmt.Println("\nYou can provide a name to use as the default sender from this profile.")
	fmt.Println("This name is optional.  It will be used as the sender name in bundled objects when")
	fmt.Println("using the default keypair for encrypting the bundle header.")
	fmt.Println("")
	defaultKPName, err := helpers.GetConsoleInputLine("Enter a default sender key name or leave empty for none")
	fmt.Println()

	if err != nil {
		return fmt.Errorf("failed while requesting the optional default sender name for the profile: %w", err)
	}

	isEncrypted := len(storeKey) > 0
	profile := &helpers.Profile{
		Name:                  profileName,
		Path:                  profilePath,
		KeyStorePath:          profileKeystorePath,
		KeyPairStorePath:      profileKeypairStorePath,
		KeyPairStoreEncrypted: isEncrypted,
		DefaultKeypairName:    defaultKPName,
	}

	fmt.Printf("Adding new profile \"%s\"...\n", profileName)
	fmt.Println()

	ch := helpers.NewConfigHelper()
	err = ch.LoadConfig()
	if err != nil {
		return fmt.Errorf("unable to load config defs: %w", err)
	}

	err = ch.NewProfile(profile)
	if err != nil {
		return fmt.Errorf("unable to add new profile to config: %w", err)
	}

	ch.Config.CurrentProfile = profileName

	fmt.Println("Saving profile definitions...")
	err = ch.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed writing profile yaml file: %w", err)
	}

	// create new keypair store and key data for this environment.
	fmt.Println("Creating new profile keypair store with default KeyPair...")
	kps, _, err := keypairs.NewKeypairStoreWithKeypair("default")
	if err != nil {
		return fmt.Errorf("failed create keypair store: %w", err)
	}

	// we set the global keypair store to our local var, so that the keystore we are
	// about to create can write out correctly, since it gets the keypair for writing from the global
	// keypair store.
	keypairs.GlobalKeyPairStore = kps

	fmt.Println("Adding new read and write keypairs for keystore...")
	_, err = kps.CreateNewKeyPair("keystore_read")
	if err != nil {
		return fmt.Errorf("unable to create keypair for keystore reads: %s", err)
	}

	_, err = kps.CreateNewKeyPair("keystore_write")
	if err != nil {
		return fmt.Errorf("unable to create keypair for keystore writes: %s", err)
	}

	fmt.Println("Saving keypair store...")
	err = kps.SaveKeyPairStore(storeKey, profileKeypairStorePath)
	if err != nil {
		return fmt.Errorf("failed save keypair store: %s", err)
	}

	// create and write out the profile keystore file
	fmt.Println("Creating keystore for new profile...")
	ks := keystore.New(profileName, "local", true)

	fmt.Println("Saving new keystore...")
	err = ks.WriteToFile(profileKeystorePath)
	if err != nil {
		return fmt.Errorf("error writing new keystore to file: %s", err)
	}

	return nil
}

// AssertEnvironmentIsEmpty is responsible for checking the primary
// config path to make sure NOTHING is there.  If something is there, whether dirs
// or files, then get permission from the user and remove all of it.
// When this returns false, it should mean that NOTHING is in this path, it's now
// a blank slate to do whatever we need to do there.
func AssertEnvironmentIsEmpty() (shouldAbort bool) {
	configPath := configdir.LocalConfig("bumblebee")
	fileSystem := os.DirFS(configPath)

	infoExists := false
	fmt.Printf("Walking: %s\n", configPath)
	_ = fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		if name == "." || name == ".." || name == ".DS_Store" {
			return nil
		}

		infoExists = true
		return errors.New("info exists")
	})

	if !infoExists {
		// there is no info in the config path, so go ahead and proceed with the init.  Do NOT abort.
		return false
	}

	fmt.Println("There are currently items in your profile's environment.  " +
		"If you proceed to init this environment, all data will be removed first.")
	fmt.Printf("\nIT MAY NOT BE RECOVERABLE AFTER REMOVING IT!!!!\n")

	inputVal, err := helpers.GetYesNoInput("Do you wish to proceed and remove any current data?", helpers.InputResponseValNo)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return true
	}

	if inputVal == helpers.InputResponseValNo {
		// The user elected NOT to remove data, so abort the init request
		return true
	}

	// The user said it's ok to remove all the data, so go ahead and do that
	if removeAllEnviromentData(configPath) != nil {
		// We were unable to remove all data, so we should abort the init
		return true
	}

	// all is good now, no need to abort
	return false
}

func removeAllEnviromentData(rootPath string) error {
	// Todo: how safe is this?  Scary stuff, Scoob!  Maybe think of a safer thing to do here in the future.
	return os.RemoveAll(rootPath)
}
