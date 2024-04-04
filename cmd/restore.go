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
	"github.com/thoughtrealm/bumblebee/bootstrap"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/symfiles"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"path/filepath"
	"strings"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores profiles from a backup file",
	Long:  "Restores profiles from a backup file",
	Run: func(cmd *cobra.Command, args []string) {
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
	fmt.Println("It is recommended that you backup your local environment before restoring profiles.")
	fmt.Println("")

	response, err := helpers.GetInputFromList(
		"Stop this restore request and backup your environment first?",
		[]helpers.InputListItem{
			{
				Option: "Y",
				Label:  "Yes, stop this restore so I can backup first",
			},
			{
				Option: "N",
				Label:  "No, do not stop. I have already backed up or do not need to do so",
			},
			{
				Option: "C",
				Label:  "Cancel this abort request",
			},
		},
		"Y",
	)

	if err != nil {
		fmt.Printf("Failure confirming backup request%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if response != "N" {
		fmt.Println("User requested to stop restore request")
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	// First, let's determine the state of our config path.  It could be empty with no
	// metadata yaml file.  We want to provide the ability to do a restore, even if that file does
	// not exist.
	err = validateConfigEnvironment()
	if err != nil {
		fmt.Printf("Failure validating local config environment%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = restoreValidateInputFile()
	if err != nil {
		fmt.Printf("Failure validating input file%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = restoreValidateKey()
	if err != nil {
		fmt.Printf("Failure validating password for backup file: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = restoreRetrieveBackupMetadata()
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

	err = updateCurrentProfile()
	if err != nil {
		fmt.Printf("Failed setting active profile%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
}

func validateConfigEnvironment() error {
	configPath, err := helpers.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed obtaining config path: %w", err)
	}

	configFilePath := filepath.Join(configPath, helpers.BBConfigFileName)

	if !helpers.FileExists(configFilePath) {
		err = initializeBasicConfigFile()
		if err != nil {
			return fmt.Errorf("failed creating a new config file: %w", err)
		}
	}

	return startBootStrap(false, false)
}

func initializeBasicConfigFile() error {
	config := &helpers.ConfigInfo{
		Profiles:       []*helpers.Profile{},
		CurrentProfile: "",
	}

	newHelper := helpers.NewConfigHelperFromConfig(config)
	return newHelper.WriteConfig()
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

	fmt.Println("")
	fmt.Println("Profiles requested for restore...")
	if len(localRestoreCommandVals.profileNames) == 0 {
		fmt.Println("** All profiles in backup file requested for restore **")
	} else {
		var missingProfileNames []string
		for idx, profileName := range localRestoreCommandVals.profileNames {
			fmt.Printf("%02d: %s\n", idx+1, profileName)

			if !restoreIsSelectedProfileInBackup(profileName) {
				missingProfileNames = append(missingProfileNames, profileName)
			}
		}

		if len(missingProfileNames) > 0 {
			fmt.Println("")
			fmt.Println("The following profiles were selected for restore, but were not found in the backup file...")
			for idx, profileName := range missingProfileNames {
				fmt.Printf("%02d: %s\n", idx+1, profileName)
			}

			fmt.Println("")

			return errors.New("some requested profiles were not found in the backup file")
		}
	}

	fmt.Println("")

	var existingProfiles []*helpers.Profile
	var profilesSelected bool
	for _, profileFromBackup := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		if !restoreIsProfileSelected(profileFromBackup) {
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

	fmt.Println("")

	if len(existingProfiles) == 0 {
		fmt.Println("None of the profiles in the backup exist locally.  They will be added as new profiles.")
		fmt.Println("Proceeding with restore...")
		return nil
	}

	fmt.Println("The following profiles were requested for restore and currently exist in the local environment...")
	for idx, profile := range existingProfiles {
		fmt.Printf("%02d: %s\n", idx+1, profile.Name)
	}

	fmt.Println("")
	fmt.Println("**Warning: These profiles will be DELETED and replaced during the restore.")
	fmt.Println("")

	response, err := helpers.GetYesNoInput("Overwrite the listed profiles?", helpers.InputResponseValNo)
	if err != nil {
		return err
	}

	if response != helpers.InputResponseValYes {
		return errors.New("user did not confirm restoring already existing profiles")
	}

	return nil
}

// restoreIsSelectedProfileInBackup will return true if the indicated profile exists in the backup file
func restoreIsSelectedProfileInBackup(profileName string) (profileExistsInBackup bool) {
	for _, profile := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		if strings.ToUpper(profile.Name) == strings.ToUpper(profileName) {
			return true
		}
	}

	return false
}

// restoreIsProfileSelected() returns true when the inputProfile exists in the list of profiles provided via
// the command line, or if no profiles are provided via the command line which indicates to restore all profiles
// that are contained in the backup file
func restoreIsProfileSelected(inputProfile *helpers.Profile) bool {
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

func restoreValidateKey() (err error) {
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

func restoreRetrieveBackupMetadata() error {
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
		if !restoreIsProfileSelected(profileFromBackup) {
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

func updateCurrentProfile() error {
	err := bootstrap.Run(false, false)
	if err != nil {
		return fmt.Errorf("error refreshing config data: %w", err)
	}

	if helpers.GlobalConfig.Config.CurrentProfile != "" {
		// current profile is set, nothing to do
		return nil
	}

	// see if we have a default profile
	defaultProfile := helpers.GlobalConfig.GetProfile("default")
	if defaultProfile != nil {
		helpers.GlobalConfig.Config.CurrentProfile = defaultProfile.Name
		err = helpers.GlobalConfig.WriteConfig()
		if err != nil {
			return fmt.Errorf("unable to update current profile to default: %w", err)
		}

		return nil
	}

	if len(helpers.GlobalConfig.Config.Profiles) == 0 {
		return errors.New("No profiles exist in config data.  Unable to set active profile.")
	}

	newCurrentProfile := helpers.GlobalConfig.Config.Profiles[0]
	helpers.GlobalConfig.Config.CurrentProfile = newCurrentProfile.Name
	err = helpers.GlobalConfig.WriteConfig()
	if err != nil {
		return fmt.Errorf("unable to update current profile to first item: %w", err)
	}

	return nil
}
