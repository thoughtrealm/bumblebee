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
	"github.com/thoughtrealm/bumblebee/streams"
	"github.com/thoughtrealm/bumblebee/symfiles"
	"github.com/vmihailenco/msgpack/v5"
	"os"
	"path/filepath"
	"strings"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Will backup one or more profiles to a symmetrically encrypted file",
	Long:  "Will backup one or more profiles to a symmetrically encrypted file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		backupProfiles(args)
	},
}

type backupCommandVals struct {
	// The user supplied key to decrypt the input with
	symmetricKey []byte

	// Command line provided symmetric key
	symmetricKeyInputText string

	// outputFile is the name of the backup file
	outputFile string

	// A slice of validated profile names
	profiles []*helpers.Profile
}

var localBackupCommandVals = &backupCommandVals{}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Use = "backup [ <profile-list> ] [flags]"
	backupCmd.Example = `  -- Backup all profiles to a file named "backup.20240328""
  bumblebee backup --output-file backup.20240328

  -- Backup the "default" profole to a file named "backup.20240328""
  bumblebee backup default --output-file default.20240328

  -- Backup profiles named "default", "bob" and "alice" to "backup.default-bob-alice.20240328" using key "supersecretkey"
  bumblebee backup default,bob,alice --output-file backup.default-bob-alice.20240328 --key supersecretkey

  -- Backup profiles named "work" and "home" to a file named "backup.work-home.20240328"
  bumblebee backup work,home --output-file backup.work-home.20240328 `

	backupCmd.Flags().StringVarP(&localBackupCommandVals.outputFile, "output-file", "y", "", "The file name to use for output")
	backupCmd.Flags().StringVarP(&localBackupCommandVals.symmetricKeyInputText, "key", "", "", "The key for encrypting the backup data. If not provided, you will be prompted for this. It is recommended to not use this value and enter via the prompt.")
}

func backupProfiles(args []string) {
	err := backupValidateProfileInfo(args)
	if err != nil {
		fmt.Printf("Failure validating profile info%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if len(localBackupCommandVals.profiles) == 1 {
		fmt.Printf("Backing up profile \"%s\"\n", localBackupCommandVals.profiles[0].Name)
	} else {
		fmt.Println("Backing up the following profiles...")
		for idx, profile := range localBackupCommandVals.profiles {
			fmt.Printf("%02d: %s\n", idx+1, profile.Name)
		}
	}

	err = backupValidateOutputfile()
	if err != nil {
		fmt.Printf("Failure validating output path%s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = backupValidateKey()
	if err != nil {
		fmt.Printf("Failure validating password for backup file: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	backupDetailsMetadata := &BackupDetailsMetadata{
		Profiles: localBackupCommandVals.profiles,
	}

	metadataBytes, err := msgpack.Marshal(backupDetailsMetadata)
	if err != nil {
		fmt.Printf("Failure preparing metadata: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	symFileWriter, err := symfiles.NewSymFileWriter(localBackupCommandVals.symmetricKey)
	if err != nil {
		fmt.Printf("Error initializing symfile writer: %s\n", helpers.FormatErrorOutputs(err))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	var inputDirs []string
	for _, profile := range localBackupCommandVals.profiles {
		inputDirs = append(inputDirs, profile.Path)
	}

	fmt.Println("Backing up profiles...")

	bytesWritten, err := symFileWriter.WriteSymFileFromDirs(inputDirs, localBackupCommandVals.outputFile, []*streams.MetadataItem{
		{Name: BACKUP_FILE_METADATA_NAME, Data: metadataBytes},
	})

	fmt.Println("Backup complete")
	fmt.Printf("%d bytes written to file\n", bytesWritten)
}

func backupValidateProfileInfo(args []string) error {
	if len(args) == 0 {
		err := backupValidateBackupAllProfiles()
		if err != nil {
			return fmt.Errorf("failure validating request to backup all profiles: %w", err)
		}
	} else {
		err := backupValidateSelectedProfiles(args)
		if err != nil {
			return fmt.Errorf("failure validating request to backup selected profiles: %w", err)
		}
	}

	if len(localBackupCommandVals.profiles) == 0 {
		return errors.New("no profiles selected for backup")
	}

	return nil
}

func backupValidateBackupAllProfiles() error {
	response, err := helpers.GetYesNoInput("Backup all profiles?", helpers.InputResponseValNo)
	if err != nil {
		return err
	}

	if response != helpers.InputResponseValYes {
		return errors.New("User did not confirm backing up all profiles")
	}

	configInfo := helpers.GlobalConfig.GetConfigInfo()

	// Really, there should never a scenario with 0 profiles, but we'll confirm just in case the environment
	// is new or hinky in some way
	if len(configInfo.Profiles) == 0 {
		return errors.New("no profiles exist for backing up")
	}

	var profileIssues []string
	var shouldAdd bool
	for _, profile := range configInfo.Profiles {
		profileIssues, shouldAdd = backupValidateProfileEntry(profile, profileIssues)
		if shouldAdd {
			localBackupCommandVals.profiles = append(localBackupCommandVals.profiles, profile)
		}
	}

	if len(profileIssues) > 0 {
		if len(profileIssues) == 1 {
			return errors.New(profileIssues[0])
		}

		return errors.New("\n  " + strings.Join(profileIssues, "\n  "))
	}

	return nil
}

func backupValidateSelectedProfiles(args []string) error {
	// we've already validated there is at least one arg, so no need to check again
	profileNamesText := args[0]
	profileNames := strings.Split(profileNamesText, ",")

	var profileIssues []string
	for _, profileName := range profileNames {
		profile := helpers.GlobalConfig.GetProfile(profileName)
		if profile == nil {
			profileIssues = append(profileIssues, fmt.Sprintf("\"%s\" does not exist", profileName))
			continue
		}

		var shouldAdd bool
		profileIssues, shouldAdd = backupValidateProfileEntry(profile, profileIssues)
		if shouldAdd {
			localBackupCommandVals.profiles = append(localBackupCommandVals.profiles, profile)
		}
	}

	if len(profileIssues) > 0 {
		if len(profileIssues) == 1 {
			return errors.New(profileIssues[0])
		}

		return errors.New("\n  " + strings.Join(profileIssues, "\n  "))
	}

	return nil
}

func backupValidateProfileEntry(aProfile *helpers.Profile, profileIssues []string) (issuesOut []string, shouldAdd bool) {
	if !helpers.DirExists(aProfile.Path) {
		profileIssues = append(profileIssues, fmt.Sprintf("Profile path for \"%s\" does not exist", aProfile.Name))
		return profileIssues, false
	}

	if !helpers.FileExists(aProfile.KeyStorePath) {
		profileIssues = append(profileIssues, fmt.Sprintf("Keystore path for \"%s\" does not exist", aProfile.Name))
		return profileIssues, false
	}

	if !helpers.FileExists(aProfile.KeyPairStorePath) {
		profileIssues = append(profileIssues, fmt.Sprintf("Keypair Store path for \"%s\" does not exist", aProfile.Name))
		return profileIssues, false
	}

	return profileIssues, true
}

func backupValidateOutputfile() error {
	if localBackupCommandVals.outputFile == "" {
		return errors.New("no --output-file specified.  --output-file is required.")
	}

	if helpers.DirExists(localBackupCommandVals.outputFile) {
		return errors.New("the provided output path refers to a path, not a file.  It must refer to a file.")
	}

	outputPath, outputFilename := filepath.Split(localBackupCommandVals.outputFile)
	if outputPath == "" {
		// if no output path is specified, should work fine, just writing to current working dir
		return nil
	}

	if outputFilename == "" {
		return errors.New("an output path was provided, but a filename was not.  The filename is required.")
	}

	if !helpers.DirExists(outputPath) {
		fmt.Printf("Path \"%s\" does not exist.\n", outputPath)
		response, err := helpers.GetYesNoInput("Create it?", helpers.InputResponseValYes)
		if err != nil {
			return fmt.Errorf("error confirming creation of new directory: %w", err)
		}

		if response == helpers.InputResponseValNo {
			return errors.New("user did not confirm creating new directory")
		}

		err = os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating new path: %w", err)
		}
	}

	return nil
}

func backupValidateKey() (err error) {
	if localBackupCommandVals.symmetricKeyInputText != "" {
		localBackupCommandVals.symmetricKey = []byte(localBackupCommandVals.symmetricKeyInputText)
		return nil
	}

	fmt.Printf("\nEnter the password for encrypting the backup file: ")
	localBackupCommandVals.symmetricKey, err = helpers.GetPasswordWithConfirm("")
	if err != nil {
		return fmt.Errorf("error entering password: %w", err)
	}

	if localBackupCommandVals.symmetricKey == nil {
		return errors.New("no password provided for backup file")
	}

	return nil
}
