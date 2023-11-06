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
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/security"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Encrypts files and secrets for transport or storage",
	Long:  "Encrypts files and secrets for transport or storage",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		bundleData()
	},
}

type bundleCommandVals struct {
	// The name of the key to use for the receiver's key. Not needed when localKeys is true
	toName string

	// The name of the sender's keypair to use.  If empty, will use the default keypair for the profile.  Not needed if localKeys is true.
	fromName string

	// If localKeys is true, then the read and write keypairs from the keypair store are used for sender and receiver.
	localKeys bool

	// inputTypeText should be console, clipboard or file
	inputTypeText string

	// inputType is transformed from inputTypeText
	inputType keystore.InputType

	// inputFilePath is the name of a file to use as input.  Only relevant for inputTypeText=file.
	inputFilePath string

	// outputTypeText should be console, clipboard or file
	outputTypeText string

	// outputType is transformed from outputTypeText
	outputType keystore.OutputType

	// outputFile is the name of a file to use as output.  Only relevant for outputTypeText=file.
	outputFile string

	// outputPath is the name of a path to use for output.  Only relevant for outputTypeText=path.
	outputPath string

	// bundleTypeText should be combined or split
	bundleTypeText string

	// bundleType is transformed from bundleTypeText
	bundleType keystore.BundleType
}

var localBundleCommandVals = &bundleCommandVals{}

type bundleSettings struct {
	receiverKey       *security.KeyInfo
	senderKey         *security.KeyInfo
	inputFile         *os.File
	cipherWriter      *cipherio.CipherWriter
	totalBytesWritten int
	// pipeBuffer        *bytes.Buffer
}

var localBundleSettings = &bundleSettings{}

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.toName, "to", "t", "", "The name of the key to use for the receiver's key data.  Not necessary if using local-keys.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.fromName, "from", "r", "", "The name of the keypair to use for the sender's key data.  If empty, uses the default keypair for the profile. Not necessary if using local-keys.")
	bundleCmd.Flags().BoolVarP(&localBundleCommandVals.localKeys, "local-keys", "l", false, "If true, will use the local store keys to write the bundle data.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputTypeText, "input-type", "i", "", "The type of the input source.  Should be one of: console, clipboard or file.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputFilePath, "input-file", "f", "", "The name of a file to use for input. Only relevant if input-type is file.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputTypeText, "output-type", "o", "", "The type of the output target.  Should be one of: console, clipboard or file.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputFile, "output-file", "y", "", "The file name to use for output. Only relevant if output-type is FILE.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputPath, "output-path", "p", "", "The path name to use for output. Only relevant if output-type is PATH.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.bundleTypeText, "bundle-type", "b", "combined", "The type of bundle to build.  Should be one of: combined or split.")
}

func bundleData() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered in bundleData(): %s\n", r)
		}
	}()

	var err error

	if localBundleCommandVals.inputTypeText == "" && localBundleCommandVals.inputFilePath != "" {
		localBundleCommandVals.inputTypeText = "file"
	}

	if localBundleCommandVals.outputTypeText == "" && localBundleCommandVals.outputFile != "" {
		localBundleCommandVals.outputTypeText = "file"
	}

	if localBundleCommandVals.outputTypeText == "" && localBundleCommandVals.outputPath != "" {
		localBundleCommandVals.outputTypeText = "path"
	}

	if localBundleCommandVals.inputTypeText == "" && helpers.CheckIsPiped() {
		localBundleCommandVals.inputTypeText = "piped"
	}

	localBundleSettings.receiverKey, localBundleSettings.senderKey, err = getKeysForBundle()
	if err != nil {
		fmt.Printf("Unable to acquire keys for bundle: %s\n", err)
		// Todo: need to update exit code for all fails
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}
	defer localBundleSettings.senderKey.Wipe()
	defer localBundleSettings.receiverKey.Wipe()

	if localBundleCommandVals.inputTypeText == "" {
		fmt.Println("No input-type provided.  --input-type is required.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	localBundleCommandVals.inputType = keystore.TextToInputType(localBundleCommandVals.inputTypeText)
	if localBundleCommandVals.inputType == keystore.InputTypeUnknown {
		fmt.Printf("Unknown input-type: \"%s\"\n", localBundleCommandVals.inputTypeText)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localBundleCommandVals.inputType == keystore.InputTypeFile {
		err = validateInputFile()
		if err != nil {
			fmt.Printf("Input file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.outputTypeText == "" {
		if !inferOutputTypeForBundle() {
			fmt.Println("Unable to infer output-type based on input-type.  You must provide a value for --output-type.")
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else {
		localBundleCommandVals.outputType = keystore.TextToOutputType(localBundleCommandVals.outputTypeText)
		if localBundleCommandVals.outputType == keystore.OutputTypeUnknown {
			fmt.Printf("Unknown output-type: \"%s\"\n", localBundleCommandVals.outputTypeText)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.outputType == keystore.OutputTypeFile {
		err = validateOutputFile()
		if err != nil {
			fmt.Printf("Output file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.outputType == keystore.OutputTypePath {
		err = validateOutputPath()
		if err != nil {
			fmt.Printf("Output path invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.bundleTypeText != "" {
		localBundleCommandVals.bundleType = keystore.TextToBundleType(localBundleCommandVals.bundleTypeText)
		if localBundleCommandVals.bundleType == keystore.BundleTypeUnknown {
			fmt.Printf("Unknown bundle type: %s\n", localBundleCommandVals.bundleTypeText)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else {
		localBundleCommandVals.bundleType = keystore.BundleTypeCombined
	}

	var totalTime time.Duration

	defer func() {
		// If input was from a file, we need to close it now, regardless of errors
		if localBundleSettings.inputFile != nil {
			_ = localBundleSettings.inputFile.Close()
		}

		if err == nil {
			p := message.NewPrinter(language.English)
			_, _ = p.Printf(
				"BUNDLE completed. Bytes written: %d in %s.\n",
				localBundleSettings.totalBytesWritten,
				helpers.FormatDuration(totalTime),
			)
		}
	}()

	localBundleSettings.cipherWriter, err = cipherio.NewCipherWriter(
		localBundleSettings.receiverKey,
		localBundleSettings.senderKey)
	if err != nil {
		fmt.Printf("Unable to create cipher writer: %s", err)
		helpers.ExitCode = helpers.ExitCodeCipherError
		return
	}
	defer localBundleSettings.cipherWriter.Wipe()

	reader, err := getInputReader()
	if err != nil {
		fmt.Printf("unable to initiate input stream: %s", err)
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	fmt.Println("Starting BUNDLE request...")
	startTime := time.Now()

	switch localBundleCommandVals.outputType {
	case keystore.OutputTypeConsole:
		err = writeToConsole(reader)
	case keystore.OutputTypeClipboard:
		err = writeToClipboard(reader)
	case keystore.OutputTypeFile:
		err = writeToFile(reader)
	case keystore.OutputTypePath:
		err = writeToPath(reader)
	default:
		// this should NEVER happen, but in case we add a new type, this will remind us during testing to call it here
		fmt.Println("Unknown output type in output writer call")
	}

	if err != nil {
		fmt.Printf("Unable to write to output stream: %s\n", err)
	}

	endTime := time.Now()
	totalTime = endTime.Sub(startTime)

	return
}

func getKeysForBundle() (receiverKI, senderKI *security.KeyInfo, err error) {
	// We will always need something from the keypair store for this so confirm it is loaded
	if keypairs.GlobalKeyPairStore == nil {
		return nil, nil, errors.New("keypair store is not loaded")
	}

	if localBundleCommandVals.localKeys {
		return getLocalKeysForBundleWrite()
	}

	if keystore.GlobalKeyStore == nil {
		return nil, nil, errors.New("keystore is not loaded")
	}

	if localBundleCommandVals.toName == "" {
		return nil, nil, errors.New("receiver key name not supplied")
	}

	// Todo: populate bundle with to and from names... get default key name from profile metadata
	// First, get the sender's keypair info
	var useSenderName = "default"
	if localBundleCommandVals.fromName != "" {
		useSenderName = localBundleCommandVals.fromName
	}

	senderKPI := keypairs.GlobalKeyPairStore.GetKeyPairInfo(useSenderName)
	if senderKPI == nil {
		return nil, nil, fmt.Errorf("Unable to locate sender's keypair for name \"%s\"\n", useSenderName)
	}
	defer senderKPI.Wipe()

	if strings.ToLower(senderKPI.Name) == "default" {
		// now that we have retrieved the sender's key, we can set the sender name to
		// something else if needed for the bundle details
		profile := helpers.GlobalConfig.GetCurrentProfile()
		if profile != nil && profile.DefaultKeypairName != "" {
			senderKPI.Name = profile.DefaultKeypairName
		} else {
			senderKPI.Name = "Not Provided"
		}
	}

	senderKI, err = security.NewKeyInfo(false, security.KeyTypeSeed, senderKPI.Name, senderKPI.Seed)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to build new sender keyinfo: %w", err)
	}

	receiverEntity := keystore.GlobalKeyStore.GetKey(localBundleCommandVals.toName)
	if receiverEntity == nil {
		return nil, nil, fmt.Errorf("receiver key not located for name \"%s\"", localBundleCommandVals.toName)
	}
	receiverKI = receiverEntity.Key

	return receiverKI.Clone(), senderKI.Clone(), nil
}

// getLocalKeysForBundleWrite will return a set of keys using the default read and write keypairs in the profile's keypair store
func getLocalKeysForBundleWrite() (receiverKeyInfo, senderKeyInfo *security.KeyInfo, err error) {
	kpiKeypairStoreRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo("keystore_read")
	if kpiKeypairStoreRead == nil {
		return nil, nil, errors.New("store default read keypair not found")
	}

	receiverPublicKey, err := kpiKeypairStoreRead.PublicKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to obtain publicKey from read keypair: %w", err)
	}

	receiverKeyInfo, err = security.NewKeyInfo(
		false,
		security.KeyTypePublic,
		"local",
		receiverPublicKey,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to build receiver key info: %w", err)
	}

	kpiKeypairStoreWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo("keystore_write")
	if kpiKeypairStoreWrite == nil {
		return nil, nil, errors.New("store default write keypair not found")
	}

	senderKeyInfo, err = security.NewKeyInfo(
		false,
		security.KeyTypeSeed,
		"local",
		[]byte(kpiKeypairStoreWrite.Seed),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to build sender key info: %w", err)
	}

	return receiverKeyInfo, senderKeyInfo, nil
}

func inferOutputTypeForBundle() (outputTypeWasInferred bool) {
	switch localBundleCommandVals.inputType {
	case keystore.InputTypeConsole:
		localBundleCommandVals.outputType = keystore.OutputTypeConsole
		return true
	case keystore.InputTypeClipboard:
		localBundleCommandVals.outputType = keystore.OutputTypeClipboard
		return false
	case keystore.InputTypeFile:
		localBundleCommandVals.outputType = keystore.OutputTypePath
		usePath, _ := filepath.Split(localBundleCommandVals.inputFilePath)
		localBundleCommandVals.outputPath = usePath
		return true
	}

	return false
}

func validateInputFile() error {
	if localBundleCommandVals.inputFilePath == "" {
		return errors.New("no input file provided")
	}

	if !helpers.FileExists(localBundleCommandVals.inputFilePath) {
		return fmt.Errorf("input file path does not exist: %s", localBundleCommandVals.inputFilePath)
	}

	return nil
}

func validateOutputFile() error {
	// inferOutputTypeForBundle will have done some of this already
	if localBundleCommandVals.outputFile == "" {
		// we will attempt to infer the output file path from the input file path
		if localBundleCommandVals.inputType != keystore.InputTypeFile {
			return errors.New("output type set to \"file\", however no output file path is provided and the input type is not \"file\", so unable to infer an output file path")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output type set to \"file\", however no output file path is provided and no input file path is provided, so unable to infer an output file path")
		}

		// an output filename with an extension of "ext" tells the cipher writer func to change the extension as needed
		localBundleCommandVals.outputFile = filepath.Join(
			localBundleCommandVals.inputFilePath, helpers.ReplaceFileExt(
				localBundleCommandVals.inputFilePath,
				".ext"))
	}

	// let's confirm that at least a filename has been provided at this point
	outputFilePath, outputFileName := filepath.Split(localBundleCommandVals.outputFile)
	if outputFileName == "" {
		// we will attempt to infer the output filename from the input file path
		if localBundleCommandVals.inputType != keystore.InputTypeFile {
			return errors.New("output type set to \"file\", however no output filename is provided and the input type is not \"file\", so unable to infer an output filename")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output type set to \"file\", however no output file path is provided and no input filename is provided, so unable to infer an output filename")
		}

		_, inputFileName := filepath.Split(localBundleCommandVals.inputFilePath)
		outputFileName = helpers.ReplaceFileExt(inputFileName, ".ext")
		localBundleCommandVals.outputFile = filepath.Join(outputFilePath, outputFileName)
	}

	// We will assume that whatever filepath has been provided or inferred is ok.  If the path doesn't exist,
	// or something else is invalid about the path, the OS will fail with an error during create and we
	// will return that to the user at that point.  We won't implement a bunch of code here to
	// reproduce that os validation behavior.

	return nil
}

// validateOutputPath will do two things...
//   - If there is no path defined yet, it will take the path from the input path.
//   - If there is a path defined, it will throw an error if the path has a file name defined, since we only want a path,
func validateOutputPath() error {
	if localBundleCommandVals.outputPath == "" {
		// we will attempt to infer the output file path from the input file path
		if localBundleCommandVals.inputType != keystore.InputTypeFile {
			return errors.New("output type set to \"file\", however no output file path is provided and the input type is not \"file\", so unable to infer an output file path")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output type set to \"file\", however no output file path is provided and no input file path is provided, so unable to infer an output file path")
		}

		inputPath, _ := filepath.Split(localBundleCommandVals.inputFilePath)
		localBundleCommandVals.outputFile = inputPath

		return nil
	}

	// a path is defined, let's validate that it does exist, and it is a path
	exists, isDir := helpers.FileExistsInfo(localBundleCommandVals.outputPath)
	if !exists {
		return fmt.Errorf("provided output path does not exist: %s", localBundleCommandVals.outputPath)
	}

	if !isDir {
		return fmt.Errorf("provided output path references a file and not a path: %s", localBundleCommandVals.outputPath)
	}

	return nil
}

func getInputReader() (io.Reader, error) {
	switch localBundleCommandVals.inputType {
	case keystore.InputTypeConsole:
		return getConsoleReader()
	case keystore.InputTypeClipboard:
		return getClipboardReader()
	case keystore.InputTypeFile:
		return getFileReader()
	case keystore.InputTypePiped:
		return getPipedReader()
	}

	return nil, errors.New("unknown input type obtaining stream reader")
}

func getConsoleReader() (io.Reader, error) {
	inputLines, err := helpers.GetConsoleMultipleInputLines("bundle")
	if err != nil {
		return nil, fmt.Errorf("unable to get user input: %s", err)
	}

	if localBundleCommandVals.outputType == keystore.OutputTypePath && localBundleCommandVals.inputFilePath == "" {
		// if an output type of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputType = keystore.OutputTypeFile
		localBundleCommandVals.outputFile = filepath.Join(localBundleCommandVals.outputPath, "bee.console.ext")
	}

	localBundleSettings.cipherWriter.OutputBundleInfo.InputSource = cipherio.BundleInputSourceDirect
	inputBytes := []byte(strings.Join(inputLines, "\n"))
	inputBuff := bytes.NewBuffer(inputBytes)
	return inputBuff, nil
}

func getClipboardReader() (io.Reader, error) {
	data, err := helpers.ReadFromClipboard()
	if err != nil {
		return nil, fmt.Errorf("unable to read from clipboard: %w", err)
	}

	reader, err := helpers.NewTextScanner(data)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize text scanner from clipboard input: %s", err)
	}

	if localBundleCommandVals.outputType == keystore.OutputTypePath && localBundleCommandVals.inputFilePath == "" {
		// if an output type of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputType = keystore.OutputTypeFile
		localBundleCommandVals.outputFile = filepath.Join(localBundleCommandVals.outputPath, "bee.clipboard.ext")
	}

	localBundleSettings.cipherWriter.OutputBundleInfo.InputSource = cipherio.BundleInputSourceDirect
	return reader, nil
}

func getPipedReader() (io.Reader, error) {
	pipeBuffer := bytes.NewBuffer(nil)
	_, err := pipeBuffer.ReadFrom(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to read piped input from stdin: %s", err)
	}

	reader, err := helpers.NewTextScanner(pipeBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to initialize pipe text scanner from pipe input: %s", err)
	}

	if localBundleCommandVals.outputType == keystore.OutputTypePath && localBundleCommandVals.inputFilePath == "" {
		// if an output type of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputType = keystore.OutputTypeFile
		localBundleCommandVals.outputFile = filepath.Join(localBundleCommandVals.outputPath, "bee.piped.ext")
	}

	localBundleSettings.cipherWriter.OutputBundleInfo.InputSource = cipherio.BundleInputSourceDirect
	return reader, nil
}

func getFileReader() (io.Reader, error) {
	f, err := os.Stat(localBundleCommandVals.inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get details from file: %s", err)
	}

	localBundleSettings.cipherWriter.OutputBundleInfo.InputSource = cipherio.BundleInputSourceFile
	localBundleSettings.cipherWriter.OutputBundleInfo.OriginalFileDate = f.ModTime().Format(time.RFC3339)

	_, name := filepath.Split(localBundleCommandVals.inputFilePath)
	localBundleSettings.cipherWriter.OutputBundleInfo.OriginalFileName = name

	file, err := os.Open(localBundleCommandVals.inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to initiate input file stream: %w", err)
	}

	localBundleSettings.inputFile = file
	return file, nil
}

func writeToConsole(reader io.Reader) (err error) {
	switch localBundleCommandVals.bundleType {
	case keystore.BundleTypeCombined:
		return writeToConsoleCombined(reader)
	case keystore.BundleTypeSplit:
		return writeToConsoleSplit(reader)
	}

	return errors.New("unknown bundle type")
}

func writeToConsoleCombined(reader io.Reader) error {
	// for now, we use line width of 32 for binary, which will be 64 chars in hex
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetConsole,
		32,
		helpers.TextWriterModeBinary,
		":start :header+data",
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	var err error
	localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToCombinedStreamFromReader(
		reader,
		textWriter,
		func(w io.Writer) error {
			var err error
			localBundleSettings.totalBytesWritten, err = textWriter.Flush()
			return err
		},
	)
	if err != nil {
		return fmt.Errorf("failed writing to console: %w", err)
	}

	return nil
}

func writeToConsoleSplit(reader io.Reader) error {
	// for now, we use line width of 32 for binary, which will be 64 chars in hex
	textWriterHdr := helpers.NewTextWriter(
		helpers.TextWriterTargetConsole,
		32,
		helpers.TextWriterModeBinary,
		":start :header",
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	textWriterData := helpers.NewTextWriter(
		helpers.TextWriterTargetConsole,
		32,
		helpers.TextWriterModeBinary,
		":start :data",
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	var err error
	localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToSplitStreamsFromReader(
		reader,
		textWriterHdr,
		textWriterData,
		func(w io.Writer) error {
			var err error
			localBundleSettings.totalBytesWritten, err = textWriterHdr.Flush()
			if err == nil {
				// for split outputs, add an extra line after the header flush only
				fmt.Println()
			}
			return err
		},
		func(w io.Writer) error {
			var err error
			localBundleSettings.totalBytesWritten, err = textWriterData.Flush()
			return err
		},
	)
	if err != nil {
		return fmt.Errorf("failed writing to console: %w", err)
	}

	return nil
}

func writeToClipboard(reader io.Reader) error {
	switch localBundleCommandVals.bundleType {
	case keystore.BundleTypeCombined:
		return writeToClipboardCombined(reader)
	case keystore.BundleTypeSplit:
		return writeToClipboardSplit(reader)
	}

	return errors.New("unknown bundle type")
}

func writeToClipboardCombined(reader io.Reader) error {
	// for now, we use line width of 32 for binary, which will be 64 chars in hex
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetClipboard,
		32,
		helpers.TextWriterModeBinary,
		":start :header+data",
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	var err error
	localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToCombinedStreamFromReader(
		reader,
		textWriter,
		func(w io.Writer) error {
			var err error
			localBundleSettings.totalBytesWritten, err = textWriter.Flush()
			return err
		},
	)

	if err != nil {
		return fmt.Errorf("failed writing to clipboard: %w", err)
	}

	return nil
}

func writeToClipboardSplit(reader io.Reader) error {
	// for now, we use line width of 32 for binary, which will be 64 chars in hex
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetClipboard,
		32,
		helpers.TextWriterModeBinary,
		":start :header",
		":end",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	var err error
	localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToSplitStreamsFromReader(
		reader,
		textWriter,
		textWriter,
		func(w io.Writer) error {
			err := textWriter.PrintFooter()
			if err != nil {
				return err
			}

			err = textWriter.PrintTextLine("")
			if err != nil {
				return err
			}

			textWriter.Reset(":start :data", ":end")
			return nil
		},
		func(w io.Writer) error {
			var err error
			localBundleSettings.totalBytesWritten, err = textWriter.Flush()
			if err == nil {
				fmt.Println("")
			}
			return err
		},
	)

	if err != nil {
		return fmt.Errorf("failed writing to clipboard: %w", err)
	}

	return nil
}

// writeToFile uses outputFile to target user provided filename
func writeToFile(reader io.Reader) error {
	var err error
	switch localBundleCommandVals.bundleType {
	case keystore.BundleTypeCombined:
		localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToCombinedFileFromReader(
			localBundleCommandVals.outputFile,
			reader)
	case keystore.BundleTypeSplit:
		localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToSplitFilesFromReader(
			localBundleCommandVals.outputFile,
			reader)
	}

	if err != nil {
		return fmt.Errorf("failed writing to file(s): %w", err)
	}

	return nil
}

// writeToPath will build out the path with a requiste filename based on input settings
func writeToPath(reader io.Reader) error {
	var err error
	switch localBundleCommandVals.bundleType {
	case keystore.BundleTypeCombined:
		_, inputFilename := filepath.Split(localBundleCommandVals.inputFilePath)
		localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToCombinedFileFromReader(
			filepath.Join(localBundleCommandVals.outputPath, helpers.ReplaceFileExt(inputFilename, ".ext")),
			reader)
	case keystore.BundleTypeSplit:
		_, inputFilename := filepath.Split(localBundleCommandVals.inputFilePath)
		localBundleSettings.totalBytesWritten, err = localBundleSettings.cipherWriter.WriteToSplitFilesFromReader(
			filepath.Join(localBundleCommandVals.outputPath, helpers.ReplaceFileExt(inputFilename, ".ext")),
			reader)
	}

	if err != nil {
		return fmt.Errorf("failed writing to file(s): %w", err)
	}

	return nil
}
