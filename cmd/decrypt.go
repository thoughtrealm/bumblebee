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
	"fmt"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/thoughtrealm/bumblebee/streams"
	"github.com/thoughtrealm/bumblebee/symfiles"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypts a file or input that was encrypted using the encrypt command and a user supplied key",
	Long:  "Decrypts a file or input that was encrypted using the encrypt command and a user supplied key",
	Run: func(cmd *cobra.Command, args []string) {
		decryptData()
	},
}

type decryptCommandVals struct {
	// The user supplied key to decrypt the input with
	symmetricKey []byte

	// Command line provided symmetric key
	symmetricKeyInputText string

	// inputSourceText should be console, clipboard, file or dirs
	inputSourceText string

	// inputSource is transformed from inputSourceText
	inputSource keystore.InputSource

	// inputFilePath is the name of a file to use as input.  Only relevant for inputSourceText=file.
	inputFilePath string

	// outputTargetText should be console, clipboard or file
	outputTargetText string

	// outputTarget is transformed from outputTargetText
	outputTarget keystore.OutputTarget

	// outputFile is the name of a file to use as output.  Only relevant for outputTargetText=file.
	outputFile string

	// outputPath is the name of a path to use for output.  Only relevant for outputTargetText=path.
	outputPath string
}

var localDecryptCommandVals = &decryptCommandVals{}

type decryptSettings struct {
	totalBytesWritten  int
	mdsr               streams.StreamReader
	symFilePayloadType symfiles.SymFilePayload
	outputFile         *os.File
	textWriter         *helpers.TextWriter
	symFileReader      symfiles.SymFileReader
	useDerivedFilename bool
}

var localDecryptSettings = &decryptSettings{}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.inputSourceText, "input-source", "i", "", "The type of the input source.  Should be one of: clipboard, file or dirs.")
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.inputFilePath, "input-file", "f", "", "The name of a file for input. Only relevant if input-source is file.")
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.outputTargetText, "output-target", "o", "", "The output target.  Should be one of: console, clipboard, file or path.")
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.outputFile, "output-file", "y", "", "The file name for output. Only relevant if output-target is FILE.")
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.outputPath, "output-path", "p", "", "The path name for output. Only relevant if output-target is FILE or PATH.")
	decryptCmd.Flags().StringVarP(&localDecryptCommandVals.symmetricKeyInputText, "key", "", "", "The key for the encrypted data. Prompted for if not provided. Prompt entry is recommended.")
}

// Todo: need to move all the validation code into a separate validateInput() func.  Same for all other relevant commands.

func decryptData() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered in decryptData(): %s\n", r)
		}
	}()

	if localDecryptCommandVals.inputFilePath != "" {
		localDecryptCommandVals.inputSourceText = "file"
	}

	if localDecryptCommandVals.outputTargetText == "" && localDecryptCommandVals.outputFile != "" {
		localDecryptCommandVals.outputTargetText = "file"
	}

	if localDecryptCommandVals.outputTargetText == "" && localDecryptCommandVals.outputPath != "" {
		localDecryptCommandVals.outputTargetText = "path"
		found, isDir := helpers.PathExistsInfo(localDecryptCommandVals.outputPath)
		if found && isDir == false {
			fmt.Printf(
				"Provided output path \"%s\" exists, but is a file, not a path\n",
				localDecryptCommandVals.outputPath)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}

		if !found {
			err := helpers.ForcePath(localDecryptCommandVals.outputPath)
			if err != nil {
				fmt.Printf(
					"Provided output path \"%s\" does not exist. An attempt to create it failed: %s\n",
					localDecryptCommandVals.outputPath,
					err)
				helpers.ExitCode = helpers.ExitCodeInvalidInput
				return
			}

			fmt.Printf("Created path \"%s\"\n", localDecryptCommandVals.outputPath)
		}
	}

	// If we are decrypting a file and no output details are provided, then we will assume
	// we want to output to a file as well.  This should result in writing the file to the
	// current directory. If the file contains a source file name stored during encryption,
	// we will use that name.  If not, we will try to derive a name from the input file name.
	if localDecryptCommandVals.inputSourceText == "file" && localDecryptCommandVals.outputTargetText == "" {
		// let's try to get away with not deriving a filename yet, let the read processor do that if we can.
		localDecryptCommandVals.outputTargetText = "file"
		localDecryptSettings.useDerivedFilename = true
	}

	// do this check after the other inference checks above relating to no supplied value for inputSourceText
	if localDecryptCommandVals.inputSourceText == "" && helpers.CheckIsPiped() {
		localDecryptCommandVals.inputSourceText = "piped"
	}

	localDecryptCommandVals.inputSource = keystore.TextToInputSource(localDecryptCommandVals.inputSourceText)
	if localDecryptCommandVals.inputSource == keystore.InputSourceUnknown {
		fmt.Println("Missing or invalid input source details.  Input details are required")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localDecryptCommandVals.outputTargetText == "" {
		if !inferOutputTargetFromInputForDecrypt() {
			fmt.Println("No output target provided and one could not be inferred from the input-source.  Please provide more explicit output  details.")
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localDecryptCommandVals.inputSource == keystore.InputSourceConsole {
		fmt.Println("Console is not a valid input target for command DECRYPT")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localDecryptCommandVals.inputFilePath != "" &&
		(localDecryptCommandVals.inputFilePath == localDecryptCommandVals.outputFile) {
		localDecryptCommandVals.outputFile = helpers.ReplaceFileExt(localDecryptCommandVals.outputFile, ".decrypted")
	}

	localDecryptCommandVals.outputTarget = keystore.TextToOutputTarget(localDecryptCommandVals.outputTargetText)
	if localDecryptCommandVals.outputTarget == keystore.OutputTargetUnknown {
		fmt.Println("Missing or invalid output details provided and none could be inferred from the input details.  Please provide output details.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localDecryptCommandVals.symmetricKeyInputText != "" {
		localDecryptCommandVals.symmetricKey = []byte(localDecryptCommandVals.symmetricKeyInputText)
	} else {
		err := getKeyForDecrypt()
		if err != nil {
			fmt.Printf("Unable to acquire data key: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if len(localDecryptCommandVals.symmetricKey) == 0 {
		// This can't really happen, but check in case anyway
		fmt.Println("No data key provided")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	defer security.Wipe(localDecryptCommandVals.symmetricKey)

	var err error

	fmt.Println("Starting Decrypt request...")
	startTime := time.Now()

	localDecryptSettings.symFileReader, err = symfiles.NewSymFileReader(
		localDecryptCommandVals.symmetricKey,
		localDecryptSettings.useDerivedFilename)

	defer localDecryptSettings.symFileReader.Wipe()

	var totalBytesWritten int
	switch localDecryptCommandVals.inputSource {
	case keystore.InputSourceFile:
		totalBytesWritten, err = decryptInputFile()
	case keystore.InputSourceConsole:
		fmt.Println("Console is not a valid input for command DECRYPT")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	case keystore.InputSourceClipboard:
		totalBytesWritten, err = decryptClipboardInput()
	case keystore.InputSourcePiped:
		totalBytesWritten, err = decryptPipedInput()
	default:
		fmt.Printf("Unknown or invalid input source: %d\n", int(localDecryptCommandVals.inputSource))
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if err != nil {
		fmt.Printf("Error decrypting data: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)

	p := message.NewPrinter(language.English)
	_, _ = p.Printf(
		"DECRYPT completed. Bytes written: %d in %s.\n",
		totalBytesWritten,
		helpers.FormatDuration(totalTime),
	)
}

func inferOutputTargetFromInputForDecrypt() bool {
	if localDecryptCommandVals.inputFilePath != "" {
		_, fileName := filepath.Split(localDecryptCommandVals.inputFilePath)
		if fileName == "" {
			return false
		}

		if filepath.Ext(fileName) == ".bsym" {
			fileName = helpers.ReplaceFileExt(fileName, ".decrypted")
		}

		localDecryptCommandVals.outputTarget = keystore.OutputTargetFile
		localDecryptCommandVals.outputTargetText = "file"
		localDecryptCommandVals.outputFile = fileName

		return true
	}

	// if we couldn't infer the output target from an input file reference, then we have nothing else to go off of
	return false
}

func getKeyForDecrypt() error {
	fmt.Printf("\nEnter a key for the encrypted data: ")
	key, err := helpers.GetPassword()
	if err != nil {
		return fmt.Errorf("unable to acquire a key for decrypting the data: %w", err)
	}

	localDecryptCommandVals.symmetricKey = bytes.Clone(key)
	security.Wipe(key)

	return nil
}

func decryptInputFile() (bytesWritten int, err error) {
	switch localDecryptCommandVals.outputTarget {
	case keystore.OutputTargetFile:
		return localDecryptSettings.symFileReader.ReadSymFile(localDecryptCommandVals.inputFilePath, localDecryptCommandVals.outputFile)
	case keystore.OutputTargetPath:
		return localDecryptSettings.symFileReader.ReadSymFile(localDecryptCommandVals.inputFilePath, localDecryptCommandVals.outputPath)
	case keystore.OutputTargetConsole:
		inputFile, err := os.Open(localDecryptCommandVals.inputFilePath)
		if err != nil {
			return 0, fmt.Errorf("unable to initialize input file: %w", err)
		}
		defer inputFile.Close()

		textWriter := helpers.NewTextWriter(
			helpers.TextWriterTargetConsole,
			32,
			helpers.TextWriterModeText,
			"",
			"",
			helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

		bytesWritten, err = localDecryptSettings.symFileReader.ReadSymReaderToWriter(inputFile, textWriter)
		if err != nil {
			return bytesWritten, fmt.Errorf("error writing sym file output to stream: %w", err)
		}

		bytesWritten, err = textWriter.Flush()
		if err != nil {
			return bytesWritten, fmt.Errorf("error finalizing output: %w", err)
		}

		return bytesWritten, err

	case keystore.OutputTargetClipboard:
		inputFile, err := os.Open(localDecryptCommandVals.inputFilePath)
		if err != nil {
			return 0, fmt.Errorf("unable to initialize input file: %w", err)
		}
		defer inputFile.Close()

		textWriter := helpers.NewTextWriter(
			helpers.TextWriterTargetClipboard,
			32,
			helpers.TextWriterModeText,
			"",
			"",
			helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

		bytesWritten, err = localDecryptSettings.symFileReader.ReadSymReaderToWriter(inputFile, textWriter)
		if err != nil {
			return bytesWritten, fmt.Errorf("error writing sym file output to stream: %w", err)
		}

		bytesWritten, err = textWriter.Flush()
		if err != nil {
			return bytesWritten, fmt.Errorf("error finalizing output: %w", err)
		}

		return bytesWritten, err
	default:
		return 0, fmt.Errorf("invalid or unknown output target: %d", err)
	}
}

func decryptClipboardInput() (bytesWritten int, err error) {
	clipboardReader, err := getClipboardReaderForDecrypt()
	if err != nil {
		return 0, fmt.Errorf("unable to initialize clipboard reader: %w", err)
	}

	return decryptReaderInput(clipboardReader)
}

func decryptPipedInput() (bytesWritten int, err error) {
	pipeReader, err := getPipedReaderForDecrypt()
	if err != nil {
		return 0, fmt.Errorf("unable to initialize pipe reader: %w", err)
	}

	return decryptReaderInput(pipeReader)
}

func decryptReaderInput(inputReader io.Reader) (bytesWritten int, err error) {
	switch localDecryptCommandVals.outputTarget {
	case keystore.OutputTargetFile:
		return localDecryptSettings.symFileReader.ReadSymReaderToFile(inputReader, localDecryptCommandVals.outputFile)
	case keystore.OutputTargetPath:
		return localDecryptSettings.symFileReader.ReadSymReaderToPath(inputReader, localDecryptCommandVals.outputPath)
	case keystore.OutputTargetConsole:
		textWriter := helpers.NewTextWriter(
			helpers.TextWriterTargetConsole,
			32,
			helpers.TextWriterModeText,
			"",
			"",
			helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

		bytesWritten, err = localDecryptSettings.symFileReader.ReadSymReaderToWriter(inputReader, textWriter)
		if err != nil {
			return bytesWritten, fmt.Errorf("error writing sym stream output to console: %w", err)
		}

		bytesWritten, err = textWriter.Flush()
		if err != nil {
			return bytesWritten, fmt.Errorf("error finalizing output: %w", err)
		}

		return bytesWritten, err

	case keystore.OutputTargetClipboard:
		textWriter := helpers.NewTextWriter(
			helpers.TextWriterTargetClipboard,
			32,
			helpers.TextWriterModeText,
			"",
			"",
			helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

		bytesWritten, err = localDecryptSettings.symFileReader.ReadSymReaderToWriter(inputReader, textWriter)
		if err != nil {
			return bytesWritten, fmt.Errorf("error writing sym stream output to clipboard: %w", err)
		}

		bytesWritten, err = textWriter.Flush()
		if err != nil {
			return bytesWritten, fmt.Errorf("error finalizing output: %w", err)
		}

		return bytesWritten, err
	default:
		return 0, fmt.Errorf("invalid or unknown output target: %d", err)
	}
}

func getClipboardReaderForDecrypt() (io.Reader, error) {
	data, err := helpers.ReadFromClipboard()
	if err != nil {
		return nil, fmt.Errorf("unable to read from clipboard: %w", err)
	}

	reader, err := helpers.NewTextScanner(data)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize text scanner from clipboard input: %s", err)
	}

	return reader, nil
}

func getPipedReaderForDecrypt() (io.Reader, error) {
	pipeBuffer := bytes.NewBuffer(nil)
	_, err := pipeBuffer.ReadFrom(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to read piped input from stdin: %s", err)
	}

	reader, err := helpers.NewTextScanner(pipeBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to initialize pipe text scanner from pipe input: %s", err)
	}

	return reader, nil
}
