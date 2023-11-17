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
	"encoding/base64"
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

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Decrypts bundled items",
	Long:  "Decrypts bundled items",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		openBundle()
	},
}

type openCommandVals struct {
	// The name of the key to use for the receiver's key. Not needed when localKeys is true
	toName string

	// The name of the sender's keypair to use.  If empty, will use the default keypair for the profile.  Not needed if localKeys is true.
	fromName string

	// If localKeys is true, then the read and write keypairs from the keypair store are used for sender and receiver.
	localKeys bool

	// inputTypeText should be clipboard, piped or file
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

	// detailsOnly just displays the details from the bundle header and input data characteristics, then exits
	detailsOnly bool

	// showAll true will display the payload key and salt when using the detailsOnly flag
	showAll bool
}

var localOpenCommandVals = &openCommandVals{}

type openSettings struct {
	receiverKey       *security.KeyPairInfo
	senderKey         *security.KeyInfo
	outputFile        *os.File
	cipherReader      *cipherio.CipherReader
	textWriter        *helpers.TextWriter
	totalBytesWritten int
}

var localOpenSettings = &openSettings{}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.Flags().StringVarP(&localOpenCommandVals.toName, "to", "t", "", "The name of the keypair to use for the receiver's key data.  If empty, uses the default keypair for the profile. Not necessary if using local-keys.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.fromName, "from", "r", "", "The name of the key to use for the sender's key data.  Not necessary if using local-keys.")
	openCmd.Flags().BoolVarP(&localOpenCommandVals.localKeys, "local-keys", "l", false, "If true, will use the local store keys to read the secret data.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.inputTypeText, "input-type", "i", "", "The type of the input source.  Should be one of: clipbloard or file.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.inputFilePath, "input-file", "f", "", "The name of a file to use for input. Only relevant if input-type is file.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.outputTypeText, "output-type", "o", "", "The type of the output target.  Should be one of: clipboard, piped or file.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.outputFile, "output-file", "y", "", "The file name to use for output. Only relevant if output-type is FILE.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.outputPath, "output-path", "p", "", "The file name to use for output. Only relevant if output-type is PATH.")
	openCmd.Flags().StringVarP(&localOpenCommandVals.bundleTypeText, "bundle-type", "b", "combined", "The type of bundle to build.  Should be one of: combined or split.")
	openCmd.Flags().BoolVarP(&localOpenCommandVals.detailsOnly, "details-only", "d", false, "Will display the bundle details only and quit. Does not extract or open the file.")
	openCmd.Flags().BoolVarP(&localOpenCommandVals.showAll, "show-all", "s", false, "True will display payload password and salt when using the details-only flag.")
}

func openBundle() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered in openBundle(): %s\n", r)
		}
	}()

	var err error

	if localOpenCommandVals.inputTypeText == "" && localOpenCommandVals.inputFilePath != "" {
		localOpenCommandVals.inputTypeText = "file"
	}

	if localOpenCommandVals.outputTypeText == "" && localOpenCommandVals.outputFile != "" {
		localOpenCommandVals.outputTypeText = "file"
	}

	if localOpenCommandVals.outputTypeText == "" && localOpenCommandVals.outputPath != "" {
		localOpenCommandVals.outputTypeText = "path"
	}

	if localOpenCommandVals.inputTypeText == "" && helpers.CheckIsPiped() {
		localOpenCommandVals.inputTypeText = "piped"
	}

	localOpenSettings.receiverKey, localOpenSettings.senderKey, err = getKeysForOpen()
	if err != nil {
		fmt.Printf("Unable to acquire keys for opening bundles: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localOpenCommandVals.inputTypeText == "" {
		fmt.Println("No input-type provided.  --input-type is required.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	localOpenCommandVals.inputType = keystore.TextToInputType(localOpenCommandVals.inputTypeText)
	if localOpenCommandVals.inputType == keystore.InputTypeUnknown {
		fmt.Printf("Unknown input-type: \"%s\"\n", localOpenCommandVals.inputTypeText)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localOpenCommandVals.inputType == keystore.InputTypeConsole {
		fmt.Println("Console input is not supported for OPEN command")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if !localOpenCommandVals.detailsOnly {
		if localOpenCommandVals.outputTypeText == "" {
			if !inferOutputTypeForOpen() {
				fmt.Println("Unable to infer output-type based on input-type.  You must provide a value for --output-type.")
				helpers.ExitCode = helpers.ExitCodeInvalidInput
				return
			}
		} else {
			localOpenCommandVals.outputType = keystore.TextToOutputType(localOpenCommandVals.outputTypeText)
			if localOpenCommandVals.outputType == keystore.OutputTypeUnknown {
				fmt.Printf("Unknown output-type: \"%s\"\n", localOpenCommandVals.outputTypeText)
				helpers.ExitCode = helpers.ExitCodeInvalidInput
				return
			}
		}
	}

	if localOpenCommandVals.bundleTypeText != "" {
		localOpenCommandVals.bundleType = keystore.TextToBundleType(localOpenCommandVals.bundleTypeText)
		if localOpenCommandVals.bundleType == keystore.BundleTypeUnknown {
			fmt.Printf("Unknown bundle type: %s\n", localOpenCommandVals.bundleTypeText)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localOpenCommandVals.inputType == keystore.InputTypeFile {
		err = validateInputFileForOpen()
		if err != nil {
			fmt.Printf("Unable to validate input file(s): %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localOpenCommandVals.outputType == keystore.OutputTypeFile || localOpenCommandVals.outputType == keystore.OutputTypePath {
		err = validateOutputFileForOpen()
		if err != nil {
			fmt.Printf("Unable to validate output file(s): %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	localOpenSettings.cipherReader, err = cipherio.NewCipherFileReader(
		localOpenSettings.receiverKey,
		localOpenSettings.senderKey)

	if err != nil {
		fmt.Printf("Error initializing cipher reader: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
	defer localOpenSettings.cipherReader.Wipe()

	var (
		writerErr    error
		outputWriter io.Writer
		totalTime    time.Duration
	)

	defer func() {
		if writerErr != nil {
			return
		}

		var deferErr error
		if localOpenSettings.outputFile != nil {
			deferErr = localOpenSettings.outputFile.Close()
			if deferErr != nil {
				fmt.Printf("Error closing output file stream: %s\n", deferErr)
			}
		}

		if localOpenSettings.textWriter != nil {
			_, deferErr = localOpenSettings.textWriter.Flush()
			if deferErr != nil {
				fmt.Printf("Error closing output console or clipboard stream: %s\n", deferErr)
			}
		}

		if err == nil {
			p := message.NewPrinter(language.English)
			if localOpenCommandVals.detailsOnly {
				_, _ = p.Printf(
					"OPEN completed in %s.\n",
					helpers.FormatDuration(totalTime))
				return
			}

			_, _ = p.Printf(
				"OPEN completed. Bytes written: %d in %s.\n",
				localOpenSettings.totalBytesWritten,
				helpers.FormatDuration(totalTime))
		}
	}()

	if localOpenCommandVals.detailsOnly {
		fmt.Println("Starting OPEN request for details only...")
	} else {
		fmt.Println("Starting OPEN request...")

		outputWriter, writerErr = getOutputWriter()
		if writerErr != nil {
			fmt.Printf("Unable to initialize output stream: %s\n", writerErr)
			return
		}
	}

	startTime := time.Now()

	switch localOpenCommandVals.inputType {
	case keystore.InputTypeConsole:
		// this should have already been denied previously, but just to be sure...
		fmt.Println("Console not allowed for input with OPEN command")
		return
	case keystore.InputTypeFile:
		err = decryptFile(outputWriter)
		if err != nil {
			fmt.Printf("Unable to decrypt input file(s): %s\n", err)
		}
	case keystore.InputTypeClipboard:
		err = decryptClipboard(outputWriter)
		if err != nil {
			fmt.Printf("Unable to decrypt clipboard input: %s\n", err)
		}
	case keystore.InputTypePiped:
		err = decryptPipe(outputWriter)
		if err != nil {
			fmt.Printf("Unable to decrypt piped input: %s\n", err)
		}
	}

	endTime := time.Now()
	totalTime = endTime.Sub(startTime)
}

func getKeysForOpen() (receiverKeyPairInfo *security.KeyPairInfo, senderKeyInfo *security.KeyInfo, err error) {
	// We will always need something from the keypair store for this so confirm it is loaded
	if keypairs.GlobalKeyPairStore == nil {
		return nil, nil, errors.New("keypair store is not loaded")
	}

	if keystore.GlobalKeyStore == nil {
		return nil, nil, errors.New("keystore is not loaded")
	}

	if localOpenCommandVals.localKeys {
		return getLocalKeysForOpenRead()
	}

	if localOpenCommandVals.fromName == "" {
		return nil, nil, errors.New("sender key name not supplied")
	}

	// First, get the receiver's keypair info
	var useReceiverName = "default"
	if localOpenCommandVals.toName != "" {
		useReceiverName = localOpenCommandVals.toName
	}

	receiverKeyPairInfo = keypairs.GlobalKeyPairStore.GetKeyPairInfo(useReceiverName)
	if receiverKeyPairInfo == nil {
		return nil, nil, fmt.Errorf("Unable to locate receiver keypair for name \"%s\"\n", useReceiverName)
	}

	senderEntity := keystore.GlobalKeyStore.GetKey(localOpenCommandVals.fromName)
	if senderEntity == nil {
		return nil, nil, fmt.Errorf("sender key not located for name \"%s\"", localOpenCommandVals.fromName)
	}
	senderKeyInfo = senderEntity.PublicKeys

	return receiverKeyPairInfo, senderKeyInfo, nil
}

// getLocalKeysForOpenRead will return a set of keys using the default read and write keypairs in the profile's keypair store
func getLocalKeysForOpenRead() (receiverKeyPairInfo *security.KeyPairInfo, senderKeyInfo *security.KeyInfo, err error) {
	kpiKeypairStoreWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiKeypairStoreWrite == nil {
		return nil, nil, errors.New("store default write keypair not found")
	}

	senderCipherPublicKey, senderSigningKey, err := kpiKeypairStoreWrite.PublicKeys()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to obtain publicKey from write keypair: %w", err)
	}

	senderKeyInfo, err = security.NewKeyInfo(
		"sender",
		senderCipherPublicKey,
		senderSigningKey,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to build sender key info: %w", err)
	}

	kpiKeypairStoreRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreReads)
	if kpiKeypairStoreRead == nil {
		return nil, nil, errors.New("store default read keypair not found")
	}

	// The reader is the receiver.  And the read key is a returned clone from GetKeyPairInfo, so ok
	// to own and return from here.
	return kpiKeypairStoreRead, senderKeyInfo, nil
}

func inferOutputTypeForOpen() (outputTypeWasInferred bool) {
	switch localOpenCommandVals.inputType {
	/* no support for console input for open command?
	case keystore.InputTypeConsole:
		localOpenCommandVals.outputType = keystore.OutputTypeConsole
		return true
	*/
	case keystore.InputTypeClipboard:
		localOpenCommandVals.outputType = keystore.OutputTypeClipboard
		return false
	case keystore.InputTypeFile:
		localOpenCommandVals.outputType = keystore.OutputTypeFile
		return true
	}

	return false
}

func validateInputFileForOpen() error {
	if localOpenCommandVals.inputFilePath == "" {
		return errors.New("input type is FILE and no input path is provided")
	}

	if localOpenCommandVals.bundleType == keystore.BundleTypeCombined {
		if !helpers.FileExists(localOpenCommandVals.inputFilePath) {
			return fmt.Errorf("input file does not exist: %s", localOpenCommandVals.inputFilePath)
		}

		if filepath.Ext(localOpenCommandVals.inputFilePath) == "" {
			helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bcomb")
		}

		return nil
	}

	// Validate split file entities
	ext := filepath.Ext(localOpenCommandVals.inputFilePath)
	if ext == "" || strings.ToLower(ext) == ".ext" {
		localOpenCommandVals.inputFilePath = helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bhdr")
	}

	if !helpers.FileExists(localOpenCommandVals.inputFilePath) {
		return fmt.Errorf("input hdr file does not exist: %s", localOpenCommandVals.inputFilePath)
	}

	bundleDataFilePath := helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bdata")
	if !helpers.FileExists(bundleDataFilePath) {
		return fmt.Errorf("input data file does not exist: %s", localOpenCommandVals.inputFilePath)
	}

	return nil
}

// validateOutputFileForOpen is called when the output type is FILE or PATH.
// The user can leave the output path or file reference empty, in which case
// we need to derive the necessary output components.
//   - When the input is  file, we want to output to the same path as the input file. In that case, the cipher writer
//
// will use the original file name from the bundle if it is available, otherwise
// it will use the input filename and change the extension.
//   - When the input is not a file, such as clipboard or console, we need to do ... something else?
func validateOutputFileForOpen() error {
	if localOpenCommandVals.inputType == keystore.InputTypeFile {
		return validateOutputFileForFileInputForOpen()
	}

	return validateOutputFileForNonFileInputsForOpen()
}

func validateOutputFileForFileInputForOpen() error {
	if localOpenCommandVals.outputFile != "" {
		return nil
	}

	var usePath string
	if localOpenCommandVals.outputPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to determine the current working directory: %w", err)
		}

		usePath = cwd
	} else {
		usePath = localOpenCommandVals.outputPath
	}

	// Force output type to PATH
	localOpenCommandVals.outputType = keystore.OutputTypePath

	// Use the input path for the output file
	localOpenCommandVals.outputPath = usePath

	return nil
}

// validateOutputFileForNonFileInputsForOpen assumes that no input file values exist
func validateOutputFileForNonFileInputsForOpen() error {
	if localOpenCommandVals.outputType == keystore.OutputTypeFile {
		if localOpenCommandVals.outputFile == "" {
			return errors.New("output type is \"file\", but no output file reference is provided")
		}

		filePath, _ := filepath.Split(localOpenCommandVals.outputFile)
		if filePath == "" {
			// should be ok with no output path, just a file name
			return nil
		}

		if !helpers.DirExists(filePath) {
			return fmt.Errorf("the output file path does not exist: %s", filePath)
		}

		return nil
	}

	if localOpenCommandVals.outputPath == "" {
		return errors.New("output type is \"path\", but no output path reference is provided")
	}

	if !helpers.DirExists(localOpenCommandVals.outputPath) {
		return fmt.Errorf("the output path does not exist: %s", localOpenCommandVals.outputPath)
	}

	return nil
}

func getOutputWriter() (io.Writer, error) {
	switch localOpenCommandVals.outputType {
	case keystore.OutputTypeConsole:
		return getConsoleWriter()
	case keystore.OutputTypeClipboard:
		return getClipboardWriter()
	case keystore.OutputTypeFile:
		return nil, nil // for type file, we pass the filename
	case keystore.OutputTypePath:
		return nil, nil // for type path, we pass the filepath
	}

	return nil, errors.New("unknown input type obtaining stream reader")
}

func getConsoleWriter() (io.Writer, error) {
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetConsole,
		32,
		helpers.TextWriterModeText,
		"",
		"",
		func() {
			fmt.Println("")
			fmt.Println("Decoded data...")
			fmt.Println("==========================================================")
		},
		func() {
			fmt.Println("==========================================================")
			fmt.Println("")
		},
	)

	localOpenSettings.textWriter = textWriter
	return textWriter, nil
}

func getClipboardWriter() (io.Writer, error) {
	textWriter := helpers.NewTextWriter(
		helpers.TextWriterTargetClipboard,
		32,
		helpers.TextWriterModeText,
		"",
		"",
		helpers.NilTextWriterEventFunc, helpers.NilTextWriterEventFunc)

	localOpenSettings.textWriter = textWriter
	return textWriter, nil
}

func getFileWriter() (io.Writer, error) {
	if localOpenCommandVals.outputFile == "" {
		localOpenCommandVals.outputFile = helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".decrypted")
	}

	var err error
	localOpenSettings.outputFile, err = os.Create(localOpenCommandVals.outputFile)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize output stream: %w", err)
	}

	return localOpenSettings.outputFile, nil
}

func decryptFile(writer io.Writer) error {
	if localOpenCommandVals.detailsOnly {
		return getBundleDetailsFromFile()
	}

	var err error
	switch localOpenCommandVals.bundleType {
	case keystore.BundleTypeCombined:
		switch localOpenCommandVals.outputType {
		case keystore.OutputTypeFile:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadCombinedFileToFile(
				localOpenCommandVals.inputFilePath,
				localOpenCommandVals.outputFile)
		case keystore.OutputTypePath:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadCombinedFileToPath(
				localOpenCommandVals.inputFilePath,
				localOpenCommandVals.outputPath)
		default:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadCombinedFileToWriter(
				localOpenCommandVals.inputFilePath,
				writer)
		}
	case keystore.BundleTypeSplit:
		// During validation, the inputFilePath extension should now point to the bunder header file
		switch localOpenCommandVals.outputType {
		case keystore.OutputTypeFile:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadSplitFilesToFile(
				localOpenCommandVals.inputFilePath,
				helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bdata"),
				localOpenCommandVals.outputFile)
		case keystore.OutputTypePath:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadSplitFilesToPath(
				localOpenCommandVals.inputFilePath,
				helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bdata"),
				localOpenCommandVals.outputPath)
		default:
			localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadSplitFilesToWriter(
				localOpenCommandVals.inputFilePath,
				helpers.ReplaceFileExt(localOpenCommandVals.inputFilePath, ".bdata"),
				writer)
		}
	}

	if err != nil {
		return fmt.Errorf("failed decrypting input file to output: %w", err)
	}

	return nil
}

// getBundleDetailsFromFile will open the input file and pass it to getBundleDetailsFromReader.
// When getting the details, the bundle type is not relevant, since the inputfile name will always
// contain the bundle header first.
func getBundleDetailsFromFile() error {
	var err error

	fileInfo, err := os.Stat(localOpenCommandVals.inputFilePath)
	if err != nil {
		return fmt.Errorf("unable to retrieve file size for details-only option: %s", err)
	}

	file, err := os.Open(localOpenCommandVals.inputFilePath)
	if err != nil {
		return fmt.Errorf("unable to open the input file: %s", err)
	}

	return getBundleDetailsFromReader(file, fileInfo.Size())
}

func decryptClipboard(writer io.Writer) error {
	if localOpenCommandVals.detailsOnly {
		return getBundleDetailsFromClipboard()
	}

	cbBytes, err := helpers.ReadFromClipboard()
	if err != nil {
		return fmt.Errorf("unable to retrieve clipboard data: %w", err)
	}

	if len(cbBytes) == 0 {
		return errors.New("no data retrieved from clipboard")
	}

	reader, err := helpers.NewTextScanner(cbBytes)
	if err != nil {
		return fmt.Errorf("unable to initialize text scanner from clipboard input: %s", err)
	}

	// Bundle type is not relevant for clipboard... all bundle types will be text encoded in with all relevant sections.
	// The text reader will parse all clipboard data into one virtual combined stream, regardless of bundle sections.
	// Since this originates from a clipboard read, it is ASSUMED that this won't be used for any huge data reads
	switch localOpenCommandVals.outputType {
	case keystore.OutputTypeFile:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadStreamToFile(
			reader,
			localOpenCommandVals.outputFile)
	case keystore.OutputTypePath:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadStreamToPath(
			reader,
			localOpenCommandVals.outputPath)
	default:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadCombinedStreamToWriter(reader, writer)
	}

	if err != nil {
		return fmt.Errorf("failed writing stream to output: %w", err)
	}

	return nil
}

func decryptPipe(writer io.Writer) error {
	pipeBuffer := bytes.NewBuffer(nil)
	_, err := pipeBuffer.ReadFrom(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read piped input from stdin: %s", err)
	}

	pbBytes := pipeBuffer.Bytes()
	if len(pbBytes) == 0 {
		return errors.New("no data returned from input pipe")
	}

	reader, err := helpers.NewTextScanner(pbBytes)
	if err != nil {
		return fmt.Errorf("unable to initialize text scanner from pipe input: %s", err)
	}

	if localOpenCommandVals.detailsOnly {
		return getBundleDetailsFromReader(reader, int64(len(pbBytes)))
	}

	// Bundle type is not relevant for piped input... all bundle types will be text encoded in with all relevant sections.
	// The text reader will parse all pipe data into one virtual combined stream, regardless of bundle sections.
	// Since this originates from a pipe read, it is ASSUMED that this won't be used for any huge data reads
	switch localOpenCommandVals.outputType {
	case keystore.OutputTypeFile:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadStreamToFile(
			reader,
			localOpenCommandVals.outputFile)
	case keystore.OutputTypePath:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadStreamToPath(
			reader,
			localOpenCommandVals.outputPath)
	default:
		localOpenSettings.totalBytesWritten, err = localOpenSettings.cipherReader.ReadCombinedStreamToWriter(reader, writer)
	}

	if err != nil {
		return fmt.Errorf("failed writing stream to output: %w", err)
	}

	return nil
}

func getBundleDetailsFromClipboard() error {
	cbBytes, err := helpers.ReadFromClipboard()
	if err != nil {
		return fmt.Errorf("unable to retrieve clipboard data: %w", err)
	}

	if len(cbBytes) == 0 {
		return errors.New("no data retrieved from clipboard")
	}

	reader, err := helpers.NewTextScanner(cbBytes)
	if err != nil {
		return fmt.Errorf("unable to initialize text scanner: %w", err)
	}

	// Bundle type is not relevant for clipboard... all bundle types will be text encoded in with all relevant sections.
	// The text reader will parse all clipboard data into one virtual combined stream, regardless of bundle sections.
	// Since this originates from a clipboard read, it is ASSUMED that this won't be used for any huge data reads
	return getBundleDetailsFromReader(reader, int64(len(cbBytes)))
}

func getBundleDetailsFromReader(r io.Reader, sourceSize int64) error {
	bundleInfo, err := localOpenSettings.cipherReader.GetBundleDetailsFromReader(r)
	if err != nil {
		return fmt.Errorf("unable to get bundle details from input stream: %w", err)
	}
	defer bundleInfo.Wipe()

	fmt.Println("")
	fmt.Println("Bundle Details")
	fmt.Println("=========================================================")
	fmt.Printf("Total Bundle Size     : %d bytes\n", sourceSize)
	fmt.Printf("Date Created          : %s\n", bundleInfo.CreateDate)
	fmt.Printf("Original File Date    : %s\n", bundleInfo.OriginalFileDate)
	fmt.Printf("Original File Name    : %s\n", bundleInfo.OriginalFileName)
	fmt.Printf("To Name               : %s\n", bundleInfo.ToName)
	fmt.Printf("From Name             : %s\n", bundleInfo.FromName)
	fmt.Printf("Input Source          : %s\n", cipherio.BundleInputSourceToText(bundleInfo.InputSource))

	if localOpenCommandVals.showAll {

		fmt.Printf("Payload Symmetric Key : %s\n", base64.RawStdEncoding.EncodeToString(bundleInfo.SymmetricKey))
		fmt.Printf("Payload RandomInput          : %s\n", base64.RawStdEncoding.EncodeToString(bundleInfo.Salt))
	}

	fmt.Println("")
	return nil
}
