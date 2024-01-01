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
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"os"
)

type importCommandVals struct {
	inputSourceText string
	inputSource     helpers.ImportInputSource
	inputFilePath   string
	password        string
	importedBytes   []byte
	nameOverride    string
	ignoreConfirm   bool
	detailsOnly     bool
}

var sharedImportCommandVals = &importCommandVals{}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [--password] [--input-source] [--input-file]",
	Short: "Imports users and keypairs from files, clipboard or piped input",
	Long:  "Imports users and keypairs from files, clipboard or piped input",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		importItem()
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&sharedImportCommandVals.inputSourceText, "input-source", "t", "", "The input source.  Should be one of: pipe, clipboard or file.")
	importCmd.Flags().StringVarP(&sharedImportCommandVals.inputFilePath, "input-file", "f", "", "The file name to use for input. Only relevant if input-source is FILE.")
	importCmd.Flags().StringVarP(&sharedImportCommandVals.nameOverride, "name", "n", "", "Overrides the name in the export package. If not provided,\nuser is prompted for name confirmation before adding to store.")
	importCmd.Flags().BoolVarP(&sharedImportCommandVals.ignoreConfirm, "ignore-confirm", "i", false, "If set, user will not be prompted to confirm the import")
	importCmd.Flags().BoolVarP(&sharedImportCommandVals.detailsOnly, "details-only", "", false, "If set, the input will not be imported, instead just the details will be displayed.\nThis allows you to validate the file before importing it.")
	importCmd.Flags().StringVarP(&sharedImportCommandVals.password,
		"password", "", "",
		`A password if required for the input stream.
If this is not provided and the input stream is password protected,
then you will be prompted for it.  Please be aware that providing 
passwords on the command line is not considered secure.
But if you are piping input or using bumblebee in a pipe/process flow, then you can
use this flag to provide passwords for input streams as needed.`)
}

// importItem flow...
//   - Validate input values.
//   - Retrieve data from input.
//   - Get password from user if required.
//   - Inform user of export package details.
//   - Get confirmation from user on import.
//   - Allow user to provide different name than indicated in the exported package.
//   - Add or update related item in keystore or keypair store.
//   - Save store to persist item change.
func importItem() {
	if inputsAreOk := validateImportInputs(); !inputsAreOk {
		// validateImportInputs() will have already printed error messages as needed
		// and set the ExitCode value.
		return
	}

	// Now that inputs are validated, we can process the import request
	var err error
	switch sharedImportCommandVals.inputSource {
	case helpers.ImportInputSourceClipboard:
		err = getImportClipboardInput()
	case helpers.ImportInputSourcePiped:
		err = getImportPipedInput()
	case helpers.ImportInputSourceFile:
		err = getImportFileInput()
	default:
		// this should never happen, but in case we forget to add support for a new input source,
		// we'll add an error output here
		logger.Errorln("Unsupported input source type detected")
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if err != nil {
		// The input funcs above will print necessary error outputs so no need to print anything here
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if len(sharedImportCommandVals.importedBytes) == 0 {
		logger.Errorln("No data was retrieved from the specified input source")
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	defer security.Wipe(sharedImportCommandVals.importedBytes)

	importProcessor := cipherio.NewImportProcessor(handleGetPasswordRequest)

	err = importProcessor.ProcessImportData(sharedImportCommandVals.importedBytes)
	if err != nil {
		logger.Errorfln("Failed processing import data: %s", err)
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	// Now, we handle the imported item
	switch importProcessor.DataType() {
	case security.ExportDataTypeKeyInfo:
		err = handleUserImport(importProcessor)
	case security.ExportDataTypeKeyPairInfo:
		err = handleKeyPairImport(importProcessor)
	case security.ExportDataTypeUnknown:
		logger.Errorln("Unknown exported data type in import data")
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	default:
		logger.Errorfln("Unsupported data type ID detected in import data: %d", int(importProcessor.DataType()))
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if err != nil {
		// The handler funcs above should have printed out the relevant error messages
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if sharedImportCommandVals.detailsOnly {
		logger.Println("Import details complete")
	} else {
		logger.Println("Import complete")
	}
}

// validateImportInputs will validate the flag values and print error messages as needed
func validateImportInputs() (inputsAreOk bool) {
	if helpers.CheckIsPiped() {
		sharedImportCommandVals.inputSourceText = "piped"
	}

	// If not piped, one of the flags --input-file or --input-source must be provided
	if sharedImportCommandVals.inputFilePath == "" && sharedImportCommandVals.inputSourceText == "" {
		logger.Errorln("No input file or input source is provided.  At least one must be provided.")
		helpers.ExitCode = helpers.ExitCodeInputError
		return false
	}

	// If an input source is provided, parse and validate that it is a valid source indicator
	if sharedImportCommandVals.inputSourceText != "" {
		sharedImportCommandVals.inputSource = helpers.TextToImportInputSource(sharedImportCommandVals.inputSourceText)
		if sharedImportCommandVals.inputSource == helpers.ImportInputSourceUnknown {
			logger.Errorfln("Unknown input source: %s", sharedImportCommandVals.inputSourceText)
			helpers.ExitCode = helpers.ExitCodeInputError
			return false
		}
	}

	// If the input source is FILE, then input-file must be provided
	if sharedImportCommandVals.inputSource == helpers.ImportInputSourceFile && sharedImportCommandVals.inputFilePath == "" {
		logger.Errorln("Input source is set to FILE, but no input file path is provided")
		helpers.ExitCode = helpers.ExitCodeInputError
		return false
	}

	// If a filepath is provided and no input source is provided, then set the input source to FILE
	if sharedImportCommandVals.inputFilePath != "" && sharedImportCommandVals.inputSourceText == "" {
		sharedImportCommandVals.inputSource = helpers.ImportInputSourceFile
	}

	// If a filepath is provided, then the input source MUST be of type FILE
	if sharedImportCommandVals.inputFilePath != "" && sharedImportCommandVals.inputSource != helpers.ImportInputSourceFile {
		logger.Errorln("An input file name was provided, but the input source was not set to FILE.")
		helpers.ExitCode = helpers.ExitCodeInputError
		return false
	}

	return true
}

func getImportClipboardInput() error {
	clipboardBytes, err := helpers.ReadFromClipboard()
	if err != nil {
		logger.Errorfln("Unable to retrieve clipboard data: %v", err)
		return err
	}

	if len(clipboardBytes) == 0 {
		return nil
	}

	sharedImportCommandVals.importedBytes, err = decodeImportedBytes(clipboardBytes)

	// decodeImportedBytes() will return an appropriate error message for returning here, if there was an error
	return err
}

func getImportPipedInput() error {
	pipeBuffer := bytes.NewBuffer(nil)
	_, err := pipeBuffer.ReadFrom(os.Stdin)
	if err != nil {
		logger.Errorfln("Unable to retrieve piped input data: %v", err)
		return err
	}

	decodedBytes, err := decodeImportedBytes(pipeBuffer.Bytes())
	if err != nil {
		// decodeImportedBytes() will return an appropriate error message for returning here
		return err
	}

	sharedImportCommandVals.importedBytes = decodedBytes
	return nil
}

func getImportFileInput() error {
	if !helpers.FileExists(sharedImportCommandVals.inputFilePath) {
		logger.Errorfln("Input file not found: %s", sharedImportCommandVals.inputFilePath)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return errors.New("failed loading file data")
	}

	file, err := os.Open(sharedImportCommandVals.inputFilePath)
	if err != nil {
		logger.Errorfln("Error opening input file: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	filebuffer := bytes.NewBuffer(nil)
	bytesRead, err := filebuffer.ReadFrom(file)
	if err != nil {
		logger.Errorfln("Error reading input file: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return err
	}

	logger.Debugfln("Bytes read from import file: %d", bytesRead)

	decodedBytes, err := decodeImportedBytes(filebuffer.Bytes())
	if err != nil {
		// decodeImportedBytes() will return an appropriate error message for returning here
		return err
	}

	sharedImportCommandVals.importedBytes = decodedBytes
	return nil
}

func decodeImportedBytes(encodedBytes []byte) (decodedBytes []byte, err error) {
	var textScanner *helpers.TextScanner
	textScanner, err = helpers.NewTextScanner(encodedBytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing imported data: %w", err)
	}

	bytesBuffer := bytes.NewBuffer(nil)
	_, err = bytesBuffer.ReadFrom(textScanner)
	if err != nil {
		return nil, fmt.Errorf("failed decoding imported data: %w", err)
	}

	return bytesBuffer.Bytes(), nil
}

func handleGetPasswordRequest() (password []byte, err error) {
	if len(sharedImportCommandVals.password) > 0 {
		return bytes.Clone([]byte(sharedImportCommandVals.password)), nil
	}

	logger.Println("A password is required to read the imported data.")
	logger.Printf("Please enter a password for the import data: ")

	password, err = helpers.GetPassword()
	logger.Println("")

	return password, err
}

func handleUserImport(importProcessor *cipherio.ImportProcessor) error {
	var importName string
	var err error

	ki := importProcessor.ImportedUser()
	if sharedImportCommandVals.detailsOnly {
		logger.Printfln("Input type        : User Public Keys")
		logger.Printfln("User Name         : %s", ki.Name)
		logger.Printfln("Cipher Public Key : %s", ki.CipherPubKey)
		logger.Printfln("Signing Public Key: %s", ki.SigningPubKey)
		logger.Println("")
		return nil
	}

	if sharedImportCommandVals.nameOverride != "" {
		importName = sharedImportCommandVals.nameOverride
	} else {
		logger.Printfln("The imported user data has the name \"%s\".", ki.Name)
		logger.Println("")
		logger.Println("Enter your preferred name for the imported data.  Leave blank to use the name from the imported data.")
		logger.Println("")

		importName, err = helpers.GetConsoleInputLine("Preferred import name (CTRL-C to cancel)")
		logger.Println("")
		if err != nil {
			logger.Errorfln("Error getting preferred import name from user: %s", err)
			helpers.ExitCode = helpers.ExitCodeRequestFailed
			return err
		}

		if importName == "" {
			importName = ki.Name
		}
	}

	kiStore := keystore.GlobalKeyStore.GetKey(importName)
	if kiStore != nil {
		if !sharedImportCommandVals.ignoreConfirm {
			// confirm updating the current user
			logger.Printfln("A user with the name \"%s\" already exists.", importName)
			logger.Println("")
			response, err := helpers.GetYesNoInput(
				fmt.Sprintf("Update info for user \"%s\" with the imported info? ", importName),
				helpers.InputResponseValNo)
			logger.Println("")

			if err != nil {
				logger.Errorfln("Error getting confirmation to update current user: %s", err)
				helpers.ExitCode = helpers.ExitCodeRequestFailed
				return err
			}

			if response == helpers.InputResponseValNo {
				logger.Errorln("User declined update")
				logger.Println("")
				helpers.ExitCode = helpers.ExitCodeRequestFailed
				return err
			}
		}

		_, err = keystore.GlobalKeyStore.UpdatePublicKeys(importName, ki.CipherPubKey, ki.SigningPubKey)
		if err != nil {
			logger.Errorfln("Unable to update keystore: %s", err)
			helpers.ExitCode = helpers.ExitCodeRequestFailed
			return err
		}

		logger.Printfln("User \"%s\" updated.", importName)
		return nil
	}

	err = keystore.GlobalKeyStore.AddKey(importName, ki.CipherPubKey, ki.SigningPubKey)
	if err != nil {
		logger.Errorfln("Error adding new user info to store: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return err
	}

	logger.Println("Keystore updated with user info")
	logger.Println("")

	return nil
}

func handleKeyPairImport(importProcessor *cipherio.ImportProcessor) error {
	return errors.New("handleKeyPairImport() not implemented")
}
