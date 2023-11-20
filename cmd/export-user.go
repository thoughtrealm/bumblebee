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

// The export-user cmd should never write out large byte sequences.  So,
// it's ok to do everything in memory, using byte buffers, []byte, etc.

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"os"
	"path/filepath"
	"strings"
)

// exportUserCmd represents the user export subcommand
var exportUserCmd = &cobra.Command{
	Use:   "user [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Exports user info for adding to another profile or system",
	Long:  "Exports user info for adding to another profile or system",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		err := startBootStrap(true, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		userName := args[0]
		exportUser(userName)
	},
}

const textOutputHeader = ":start  :export-user  :hex"

var exportUserFromKeypair bool

func init() {
	exportCmd.AddCommand(exportUserCmd)
	exportUserCmd.Flags().BoolVarP(
		&exportUserFromKeypair, "from-keypair", "", false,
		`Extracts only public keys from a keypair and exports as a user,
which is safe to send to other users. Keypair is located using provided username.
If the provided username is "default", then the following occurs...
-- The default sender name for the profile will be checked and used for the exported default user info.  
-- If no default sender name exists, then you will be prompted for the name to use
instead of default.
If no keypair is located with the provided username, then the default sender name 
will be checked and used if it matches the provided name.`)
}

func exportUser(userName string) {
	var err error
	var entity *security.Entity

	logger.Println("")
	logger.Debugfln("Using profile: %s\n", helpers.GlobalConfig.GetCurrentProfile().Name)

	if exportUserFromKeypair {
		entity, err = getExportEntityFromKeyPair(userName)
		if err != nil {
			logger.Printfln("Unable to obtain user info from indicated keypair: %v", err)
			return
		}
	} else {
		entity = keystore.GlobalKeyStore.GetKey(userName)
		if entity == nil {
			logger.Errorfln("No user with the name \"%s\" was found.\n", userName)
			return
		}
	}

	var passwordBytes []byte
	if sharedExportCommandVals.exportPassword != "" {
		passwordBytes = []byte(sharedExportCommandVals.exportPassword)
	} else {
		logger.Println(
			"If you would like, you can provide an optional password for securing the export file. " +
				"This is not required for sharing public keys, but does offer an extra level of secrecy when " +
				"required.")
		logger.Println("")
		logger.Println(
			"If you choose to provide a password, you will have to provide that password to whomever " +
				"will be importing the file.")
		logger.Println("")
		logger.Printf("Enter an optional password or leave empty for no password: ")

		passwordBytes, err = helpers.GetPasswordWithConfirm("password")
		logger.Println("")
		if err != nil {
			logger.Errorfln("Failed trying to obtain password for export file: %s", err)
			helpers.ExitCode = helpers.ExitCodeRequestFailed
			return
		}
	}

	eki, _ := security.NewExportKeyInfoFromKeyInfo(entity.PublicKeys)

	sharedProcessExportFlags()

	if sharedExportCommandVals.exportOutputFilePath != "" {
		sharedExportCommandVals.exportOutputTarget = helpers.ExportOutputTargetFile
	}

	exportWriter := cipherio.NewExportWriter(passwordBytes)
	defer exportWriter.Wipe()

	switch sharedExportCommandVals.exportOutputTarget {
	case helpers.ExportOutputTargetConsole:
		err = exportUserInfoToConsole(exportWriter, passwordBytes, eki)
	case helpers.ExportOutputTargetClipboard:
		err = exportUserInfoToClipboard(exportWriter, passwordBytes, eki)
	case helpers.ExportOutputTargetFile:
		err = exportUserInfoToFile(exportWriter, passwordBytes, eki)
	case helpers.ExportOutputTargetUnknown:
		err = fmt.Errorf("unknown export output target: %s", sharedExportCommandVals.exportOutputTargetText)
	}

	if err != nil {
		logger.Errorfln("Export failed: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
}

func getExportEntityFromKeyPair(userName string) (entity *security.Entity, err error) {
	logger.Debugfln("Building export key info from keypair using username: %s", userName)

	var defaultKeypairName string
	profile := helpers.GlobalConfig.GetCurrentProfile()
	if profile != nil && profile.DefaultKeypairName != "" {
		defaultKeypairName = profile.DefaultKeypairName
	}

	tempUserName := userName
	kpi := keypairs.GlobalKeyPairStore.GetKeyPairInfo(tempUserName)
	if kpi != nil {
		logger.Debugfln("KeyPair located using username: %s", tempUserName)
	} else {
		logger.Debug("Unable to locate a keypair using the provided username.  Checking if provided username is the default profile name")
		if !strings.EqualFold(tempUserName, defaultKeypairName) {
			logger.Debug("The username provided is NOT the default profile name.  Can't find KPI.")
			return nil, fmt.Errorf("unable to locate a keypair using username \"%s\" and the provided username is not the default profile name", userName)
		}

		tempUserName = defaultKeypairName
		kpi = keypairs.GlobalKeyPairStore.GetKeyPairInfo("default")
		if kpi == nil {
			// Really, this should never happen. The default keypair should always be available.
			// Something must be wrong with the keypair store?
			return nil, fmt.Errorf("unable to locate a keypair using username \"%s\" or the system keypair named \"default\"", userName)
		}

		logger.Debugfln("Using profile's DefaultKeypairName as username: %s", defaultKeypairName)
	}

	if strings.EqualFold(tempUserName, "default") && defaultKeypairName != "" {
		// This should only be the case if they passed in the username "default"
		tempUserName = defaultKeypairName
	}

	// One more final check, just in case they did something like set the DefaultKeyPair name to "default"
	if strings.EqualFold(tempUserName, "default") && defaultKeypairName != "" {
		logger.Println("No appropriate default keypair name was found.")
		tempUserName, err = helpers.GetConsoleInputLine("Please enter your desired export username (CTRL-C to CANCEL):")
		if err != nil {
			return nil, fmt.Errorf("failed while obtaining appropriate export username: %w", err)
		}
	}

	cipherPubKey, signingPubKey, err := kpi.PublicKeys()
	if err != nil {
		return nil, fmt.Errorf("unable to extract public keys from retrieved keypair info: %w", err)
	}

	ki, _ := security.NewKeyInfo(
		tempUserName,
		cipherPubKey,
		signingPubKey,
	)

	entityOut := &security.Entity{
		Name:       tempUserName,
		PublicKeys: ki,
	}

	return entityOut, nil
}

func exportUserInfoToConsole(exportWriter *cipherio.ExportWriter, password []byte, eki *security.ExportKeyInfo) error {
	if sharedExportCommandVals.exportOutputEncoding == helpers.ExportOutputEncodingRaw {
		return errors.New("output encoding \"BINARY\" is not compatible with output target \"CONSOLE\"")
	}

	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetConsole,
		32,
		helpers.TextWriterModeBinary,
		textOutputHeader,
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	err := exportWriter.WriteExportKeyInfoToStream(eki, password, textWriter)
	if err != nil {
		return fmt.Errorf("unable to write to console: %w", err)
	}

	_, err = textWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed finalizing stream output to console: %w", err)
	}

	// For console output, we need another empty line
	logger.Println("")
	logger.Println("Export complete")
	return nil
}

func exportUserInfoToClipboard(exportWriter *cipherio.ExportWriter, password []byte, eki *security.ExportKeyInfo) error {
	if sharedExportCommandVals.exportOutputEncoding == helpers.ExportOutputEncodingRaw {
		return errors.New("output encoding \"RAW\" is not compatible with output target \"CLIPBOARD\"")
	}

	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetClipboard,
		32,
		helpers.TextWriterModeBinary,
		textOutputHeader,
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	err := exportWriter.WriteExportKeyInfoToStream(eki, password, textWriter)
	if err != nil {
		return fmt.Errorf("unable to write to clipboard: %w", err)
	}

	_, err = textWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed finalizing stream output to clipboard: %w", err)
	}

	fmt.Println("Data exported to clipboard.")
	return nil
}

func exportUserInfoToFile(exportWriter *cipherio.ExportWriter, password []byte, eki *security.ExportKeyInfo) error {
	if sharedExportCommandVals.exportOutputFilePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to determine the current working directory: %s", err)
		}

		if eki.Name == "" {
			return errors.New("cannot determine output file name: no filepath provided and user info contains no name")
		}

		var ext string
		switch sharedExportCommandVals.exportOutputEncoding {
		case helpers.ExportOutputEncodingRaw:
			ext = ".bexp-dat"
		case helpers.ExportOutputEncodingText:
			ext = ".bexp-txt"
		}

		sharedExportCommandVals.exportOutputFilePath = filepath.Join(
			cwd,
			helpers.GetFileSafeName(eki.Name)+ext)
	}

	switch sharedExportCommandVals.exportOutputEncoding {
	case helpers.ExportOutputEncodingRaw:
		return exportUserInfoToFileWithRawEncoding(exportWriter, password, eki)
	case helpers.ExportOutputEncodingText:
		return exportUserInfoToFileWithTextEncoding(exportWriter, password, eki)
	default:
		return errors.New("unknown output encoding for exporting key info to file")
	}
}

func exportUserInfoToFileWithRawEncoding(exportWriter *cipherio.ExportWriter, password []byte, eki *security.ExportKeyInfo) error {
	err := exportWriter.WriteExportKeyInfoToFile(eki, password, sharedExportCommandVals.exportOutputFilePath)
	if err != nil {
		return fmt.Errorf("unable to export info to file: %w", err)
	}

	logger.Printfln("Export info written to file: %s\n", sharedExportCommandVals.exportOutputFilePath)
	logger.Println("File export complete")
	return nil
}

func exportUserInfoToFileWithTextEncoding(exportWriter *cipherio.ExportWriter, password []byte, eki *security.ExportKeyInfo) error {
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetBuffered,
		32,
		helpers.TextWriterModeBinary,
		textOutputHeader,
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	err := exportWriter.WriteExportKeyInfoToStream(eki, password, textWriter)
	if err != nil {
		return fmt.Errorf("unable to write info to stream buffer: %w", err)
	}

	_, err = textWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed finalizing stream output to buffer: %w", err)
	}

	exportBytes := textWriter.PostFlushOutputBuffer()

	outputFile, err := os.Create(sharedExportCommandVals.exportOutputFilePath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	_, err = outputFile.Write(exportBytes)
	if err != nil {
		return fmt.Errorf("unable to write data to rfile: %w", err)
	}

	logger.Printfln("Export info written to file: %s\n", sharedExportCommandVals.exportOutputFilePath)
	logger.Println("File export complete")
	return nil
}
