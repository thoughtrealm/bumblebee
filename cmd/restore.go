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

	fmt.Println("Restore completed")
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

func backupExecuteRestore() error {
	var (
		includePaths     []string
		profilesSelected bool
		err              error
	)

	// ** Get the target path here?

	for _, profileFromBackup := range localRestoreCommandVals.backupDetailsMetadata.Profiles {
		if !backupProfileIsSelected(profileFromBackup) {
			continue
		}

		includePaths, err = backupAddIncludePathFromProfile(profileFromBackup, includePaths)
		if err != nil {
			return fmt.Errorf("failed adding include path from profile: %w", err)
		}

		profilesSelected = true
		if backupProfileExistsLocally(profileFromBackup) {
			err = backupEmptyProfilePath(profileFromBackup)
			if err != nil {
				return fmt.Errorf("failed emptying local profile path: %w", err)
			}
		}
	}

	if !profilesSelected {
		return errors.New("no profiles in the backup were selected for restoring.  Nothing to restore.")
	}

	return nil
}

func backupEmptyProfilePath(profile *helpers.Profile) error {

}

func backupAddIncludePathFromProfile(profile *helpers.Profile, includePaths []string) ([]string, error) {

}
