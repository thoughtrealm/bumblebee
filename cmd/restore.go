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

package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/symfiles"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"strings"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores profiles from a backup file",
	Long:  "Restores profiles from a backup file",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		restoreProfiles(args)
	},
}

type restoreCommandVals struct {
	// The user supplied key to decrypt the input with
	symmetricKey []byte

	// Command line provided symmetric key
	symmetricKeyInputText string

	// inputFile is the name of the backup file
	inputFile string

	// A slice of profile names
	profileNames []string

	// The BackupDetailsMetadata struct read from the backup file
	backupDetailsMetadata *BackupDetailsMetadata

	// The totalBytesWritten returned from the symFile reader/writer
	totalByteswritten int
}

var localRestoreCommandVals = &restoreCommandVals{}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Use = "restore [ <profile-list> ] [flags]"
	restoreCmd.Example = `  -- Restore all profiles in the backup file named "backup.20240328.bsym"
  bumblebee restore --input-file backup.20240328

  -- Restorew the "default" profile in the backup file named "backup.20240328"
  bumblebee restore default --input-file default.20240328

  -- Restore profiles named "default", "bob" and "alice" from a backup file "backup.default-bob-alice.20240328" using key "supersecretkey"
  bumblebee restore default,bob,alice --input-file backup.default-bob-alice.20240328 --key supersecretkey

  -- Restore profiles named "work" and "home" from a backup file file named "backup.work-home.20240328"
  bumblebee restore work,home --input-file backup.work-home.20240328 `

	restoreCmd.Flags().StringVarP(&localRestoreCommandVals.inputFile, "input-file", "i", "", "The file name to use for input.")
	restoreCmd.Flags().StringVarP(&localRestoreCommandVals.symmetricKeyInputText, "key", "", "", "The key for decrypting the backup data. If not provided, you will be prompted for this. It is recommended to not use this value and enter via the prompt.")
}

func restoreProfiles(args []string) {
	err := restoreValidateInputFile()
	if err != nil {
		fmt.Printf("Failure validating input file%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = backupRestoreValidateKey()
	if err != nil {
		fmt.Printf("Failure validating password for backup file: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = backupRetrieveBackupMetadata()
	if err != nil {
		fmt.Printf("Failure retrieving metadata from backup file: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("")
	fmt.Println("Profiles contained in backup file...")
	for idx, profile := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		fmt.Printf("%02d: %s\n", idx+1, profile.Name)
	}
	fmt.Println("")

	err = restoreValidateInputProfiles(args)
	if err != nil {
		fmt.Printf("Failure parsing profiles%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = backupExecuteRestore()
	if err != nil {
		fmt.Printf("Restore failed%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Printf("Restore completed. Total bytes written: %d\n", localRestoreCommandVals.totalByteswritten)
}

func restoreValidateInputFile() error {
	if localRestoreCommandVals.inputFile == "" {
		return errors.New("No value provided for --input-file.  --input-file is required.")
	}

	if !helpers.FileExists(localRestoreCommandVals.inputFile) {
		return errors.New(fmt.Sprintf("input file \"%s\" does not exist", localRestoreCommandVals.inputFile))
	}

	return nil
}

func restoreValidateInputProfiles(args []string) error {
	if len(args) > 0 {
		inputProfileNamesText := args[0]
		localRestoreCommandVals.profileNames = strings.Split(inputProfileNamesText, ",")
	}

	var existingProfiles []*helpers.Profile
	var profilesSelected bool
	for _, profileFromBackup := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		if !backupProfileIsSelected(profileFromBackup) {
			continue
		}

		profilesSelected = true
		if backupProfileExistsLocally(profileFromBackup) {
			existingProfiles = append(existingProfiles, profileFromBackup)
		}
	}

	if !profilesSelected {
		return errors.New("No requested profiles exist in the backup.  Nothing to restore.")
	}

	if len(existingProfiles) == 0 {
		return nil
	}

	fmt.Println("The following profiles will be restored and currently exist in the local environment...")
	for idx, profile := range existingProfiles {
		fmt.Printf("%02d: %s\n", idx+1, profile.Name)
	}

	fmt.Println("")
	fmt.Println("**Warning: These profiles will be DELETED and replaced during the restore.")
	fmt.Println("")

	response, err := helpers.GetYesNoInput("Overwrite the listed profiles?", helpers.InputResponseValNo)
	fmt.Println("")
	if err != nil {
		return err
	}

	if response != helpers.InputResponseValYes {
		return errors.New("user did not confirm restoring already existing profiles")
	}

	return nil
}

// backupProfileIsSelected() returns true when the inputProfile exists in the list of profiles provided via
// the command line, or if no profiles are provided via the command line which indicates to restore all profiles
// that are contained in the backup file
func backupProfileIsSelected(inputProfile *helpers.Profile) bool {
	if inputProfile == nil {
		return false
	}

	if len(localRestoreCommandVals.profileNames) == 0 {
		return true
	}

	inputProfileNameToUpper := strings.ToUpper(inputProfile.Name)
	inputProfileAliasToUpper := strings.ToUpper(inputProfile.DefaultKeypairName)
	for _, profileName := range localRestoreCommandVals.profileNames {
		profileNameToUpper := strings.ToUpper(profileName)
		if inputProfileNameToUpper == profileNameToUpper || inputProfileAliasToUpper == profileNameToUpper {
			return true
		}
	}

	return false
}

func backupProfileExistsLocally(inputProfile *helpers.Profile) bool {
	if helpers.GlobalConfig.GetProfile(inputProfile.Name) != nil {
		return true
	}

	return false
}

func backupRestoreValidateKey() (err error) {
	if localRestoreCommandVals.symmetricKeyInputText != "" {
		localRestoreCommandVals.symmetricKey = []byte(localRestoreCommandVals.symmetricKeyInputText)
		return nil
	}

	fmt.Printf("\nEnter the password for decrypting the backup file: ")
	localRestoreCommandVals.symmetricKey, err = helpers.GetPassword()
	if err != nil {
		return fmt.Errorf("error entering password: %w", err)
	}

	if localRestoreCommandVals.symmetricKey == nil {
		return errors.New("no password provided for backup file")
	}

	return nil
}

func backupRetrieveBackupMetadata() error {
	symFileReader, err := symfiles.NewSymFileReader(localRestoreCommandVals.symmetricKey, false, nil)
	if err != nil {
		return fmt.Errorf("failure initializing symfile reader: %w\n", err)
	}

	mc, err := symFileReader.ReadSymFileMetadata(localRestoreCommandVals.inputFile)
	if err != nil {
		return fmt.Errorf("failure retrieving metadata from backup file: %w", err)
	}

	metadataProfileDataItem := mc.GetMetadataItem(BACKUP_FILE_METADATA_NAME)
	if metadataProfileDataItem == nil {
		return fmt.Errorf("backup metadata not found in backup file: %w", err)
	}

	err = msgpack.Unmarshal(metadataProfileDataItem.Data, &localRestoreCommandVals.backupDetailsMetadata)
	if err != nil {
		return fmt.Errorf("failure unmarshaling metadata: %w", err)
	}

	return nil
}

// backupExecuteRestore will execute the restore functionality.  By the time this is called,
// all required validations and initializations are complete.
func backupExecuteRestore() error {
	/*
		Pseudologic...
		** Note: We will only support local profile environment first.  Later, we will support remote or custom paths.
		1- Iterate through the profiles in the backup file.
		2- Ignore ones that are not selected and do not add them to the includePaths list.
		3- Selected profiles should be added to the includePaths list.
		4- If already exists in the local setup, then remove the local profile path and let the restore recreate it.
		5- If doesn't exist locally, add the new profile to the config and store the changed yaml file.
		6- Run the symfile to path functionality, passing in the includePaths to only restore selected profiles.
	*/

	var (
		includePaths     []string
		profilesSelected bool
		err              error
	)

	configPath, err := helpers.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed retrieving local config path: %w", err)
	}

	for _, profileFromBackup := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		if !backupProfileIsSelected(profileFromBackup) {
			continue
		}

		var profileDirName, profilePath, profileKeystorePath, profileKeypairStorePath string
		profileDirName, profilePath, profileKeystorePath, profileKeypairStorePath, err = helpers.GetNewProfilePaths(profileFromBackup.Name)
		if err != nil {
			return fmt.Errorf("failed building profile paths: %w", err)
		}

		includePaths = append(includePaths, profileDirName)
		profilesSelected = true

		if backupProfileExistsLocally(profileFromBackup) {
			err = os.RemoveAll(profilePath)
			if err != nil {
				return fmt.Errorf("failed removing current local path")
			}

			//
			continue
		}

		// This is a new profile, so create a new profile and add it to the config
		newProfile := &helpers.Profile{
			Name:                  profileFromBackup.Name,
			Path:                  profilePath,
			KeyStorePath:          profileKeystorePath,
			KeyPairStorePath:      profileKeypairStorePath,
			KeyPairStoreEncrypted: profileFromBackup.KeyPairStoreEncrypted,
			DefaultKeypairName:    profileFromBackup.DefaultKeypairName,
		}

		configHelper := helpers.NewConfigHelper()
		err = configHelper.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed loading config defs: %w", err)
		}

		err = configHelper.NewProfile(newProfile)
		if err != nil {
			return fmt.Errorf("failed adding new profile to config: %w", err)
		}

		err = configHelper.WriteConfig()
		if err != nil {
			return fmt.Errorf("failed writing profile config file: %w", err)
		}
	}

	if !profilesSelected || len(includePaths) == 0 {
		return errors.New("no profiles in the backup were selected for restoring.  Nothing to restore.")
	}

	symFileReader, err := symfiles.NewSymFileReader(
		localRestoreCommandVals.symmetricKey, false, includePaths)
	if err != nil {
		return fmt.Errorf("error creating symFileReader: %w", err)
	}

	file, err := os.Open(localRestoreCommandVals.inputFile)
	if err != nil {
		return fmt.Errorf("failed opening backup file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	localRestoreCommandVals.totalByteswritten, err = symFileReader.ReadSymReaderToPath(file, configPath)
	if err != nil {
		return fmt.Errorf("failed restoring profiles: %w", err)
	}

	return nil
}
