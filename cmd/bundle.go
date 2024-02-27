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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/thoughtrealm/bumblebee/streams"
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

	// inputSourceText should be console, clipboard, file or dirs
	inputSourceText string

	// inputSource is transformed from inputSourceText
	inputSource keystore.InputSource

	// inputFilePath is the name of a file to use as input.  Only relevant for inputSourceText=file.
	inputFilePath string

	// inputDir is the name of a directory to use as input.  Only relevant for inputSourceText=dirs.
	inputDir string

	// inputDescriptorPath is a path to a file that contains a list of input paths.  Only relevant for inputSourceText=dirs.
	inputDescriptorPath string

	// outputTargetText should be console, clipboard or file
	outputTargetText string

	// outputTarget is transformed from outputTargetText
	outputTarget keystore.OutputTarget

	// outputFile is the name of a file to use as output.  Only relevant for outputTargetText=file.
	outputFile string

	// outputPath is the name of a path to use for output.  Only relevant for outputTargetText=path.
	outputPath string

	// bundleTypeText should be combined or split
	bundleTypeText string

	// bundleType is transformed from bundleTypeText
	bundleType keystore.BundleType
}

var localBundleCommandVals = &bundleCommandVals{}

type bundleSettings struct {
	receiverKI        *security.KeyInfo
	senderKPI         *security.KeyPairInfo
	inputFile         *os.File
	cipherWriter      *cipherio.CipherWriter
	totalBytesWritten int
	mdsr              streams.StreamReader
	// pipeBuffer        *bytes.Buffer
}

var localBundleSettings = &bundleSettings{}

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.toName, "to", "t", "", "The name of the key to use for the receiver's key data.  Not necessary if using local-keys.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.fromName, "from", "r", "", "The name of the keypair to use for the sender's key data.  If empty, uses the default keypair for the profile. Not necessary if using local-keys.")
	bundleCmd.Flags().BoolVarP(&localBundleCommandVals.localKeys, "local-keys", "l", false, "If true, will use the local store keys to write the bundle data.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputSourceText, "input-source", "i", "", "The type of the input source.  Should be one of: console, clipboard, file or dirs.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputFilePath, "input-file", "f", "", "The name of a file to use for input. Only relevant if input-source is file.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputDir, "input-dir", "", "", "The name of a directory to use for input. Only relevant if input-source is dirs.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.inputDescriptorPath, "input-descriptor", "", "", "The name of a file that contains a list of directories for input. Only relevant if input-source is dirs.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputTargetText, "output-target", "o", "", "The output target.  Should be one of: console, clipboard or file.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputFile, "output-file", "y", "", "The file name to use for output. Only relevant if output-target is FILE.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.outputPath, "output-path", "p", "", "The path name to use for output. Only relevant if output-target is PATH.")
	bundleCmd.Flags().StringVarP(&localBundleCommandVals.bundleTypeText, "bundle-type", "b", "combined", "The type of bundle to build.  Should be one of: combined or split.")
}

func bundleData() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered in bundleData(): %s\n", r)
		}
	}()

	var err error

	if localBundleCommandVals.inputSourceText == "" {
		if localBundleCommandVals.inputFilePath != "" {
			localBundleCommandVals.inputSourceText = "file"
		} else if localBundleCommandVals.inputDir != "" {
			localBundleCommandVals.inputSourceText = "dirs"
		} else if localBundleCommandVals.inputDescriptorPath != "" {
			localBundleCommandVals.inputSourceText = "dirs"
		}
	}

	if localBundleCommandVals.outputTargetText == "" && localBundleCommandVals.outputFile != "" {
		localBundleCommandVals.outputTargetText = "file"
	}

	if localBundleCommandVals.outputTargetText == "" && localBundleCommandVals.outputPath != "" {
		localBundleCommandVals.outputTargetText = "path"
	}

	// do this check after the other inference checks above relating to no supplied value for inputSourceText
	if localBundleCommandVals.inputSourceText == "" && helpers.CheckIsPiped() {
		localBundleCommandVals.inputSourceText = "piped"
	}

	localBundleSettings.receiverKI, localBundleSettings.senderKPI, err = getKeysForBundle()
	if err != nil {
		fmt.Printf("Unable to acquire keys for bundle: %s\n", err)
		// Todo: need to update exit code for all fails
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}
	defer localBundleSettings.senderKPI.Wipe()

	if localBundleCommandVals.inputSourceText == "" {
		fmt.Println("No input-source provided.  --input-source is required.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	localBundleCommandVals.inputSource = keystore.TextToInputSource(localBundleCommandVals.inputSourceText)
	if localBundleCommandVals.inputSource == keystore.InputSourceUnknown {
		fmt.Printf("Unknown input-source: \"%s\"\n", localBundleCommandVals.inputSourceText)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localBundleCommandVals.inputSource == keystore.InputSourceFile {
		err = validateInputFile()
		if err != nil {
			fmt.Printf("Input file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.inputSource == keystore.InputSourceDirs {
		err = validateInputDirs()
		if err != nil {
			fmt.Printf("Input dirs invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.outputTargetText == "" {
		if !inferOutputTargetForBundle() {
			fmt.Println("Unable to infer output-target based on input-source.  You must provide a value for --output-target.")
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else {
		localBundleCommandVals.outputTarget = keystore.TextToOutputTarget(localBundleCommandVals.outputTargetText)
		if localBundleCommandVals.outputTarget == keystore.OutputTargetUnknown {
			fmt.Printf("Unknown output-target: \"%s\"\n", localBundleCommandVals.outputTargetText)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.inputSource == keystore.InputSourceDirs &&
		localBundleCommandVals.outputTarget != keystore.OutputTargetFile {
		fmt.Println("Incorrect output target for input source DIRS.  Output target MUST BE of type FILE.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localBundleCommandVals.outputTarget == keystore.OutputTargetFile {
		err = validateOutputFile()
		if err != nil {
			fmt.Printf("Output file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localBundleCommandVals.outputTarget == keystore.OutputTargetPath {
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
		localBundleSettings.receiverKI,
		localBundleSettings.senderKPI)
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

	switch localBundleCommandVals.outputTarget {
	case keystore.OutputTargetConsole:
		err = writeToConsole(reader)
	case keystore.OutputTargetClipboard:
		err = writeToClipboard(reader)
	case keystore.OutputTargetFile:
		err = writeToFile(reader)
	case keystore.OutputTargetPath:
		err = writeToPath(reader)
	default:
		// this should NEVER happen, but in case we add a new type, this will remind us during testing to call it here
		fmt.Println("Unknown output target in output writer call")
	}

	if err != nil {
		fmt.Printf("Unable to write to output stream: %s\n", err)
	}

	endTime := time.Now()
	totalTime = endTime.Sub(startTime)

	return
}

func getKeysForBundle() (receiverKeyInfo *security.KeyInfo, senderKeyPairInfo *security.KeyPairInfo, err error) {
	// We will always need something from the keypair store for this so confirm it is loaded
	if keypairs.GlobalKeyPairStore == nil {
		return nil, nil, errors.New("keypair store is not loaded")
	}

	if keystore.GlobalKeyStore == nil {
		return nil, nil, errors.New("keystore is not loaded")
	}

	if localBundleCommandVals.localKeys {
		return getLocalKeysForBundleWrite()
	}

	if localBundleCommandVals.toName == "" {
		return nil, nil, errors.New("receiver key name not supplied")
	}

	// First, get the sender's keypair info
	var useSenderName = "default"
	if localBundleCommandVals.fromName != "" {
		useSenderName = localBundleCommandVals.fromName
	}

	// GetKeyPairInfo returns cloned value.  It can be wiped later, but for now we do not wipe,
	// since we are passing it back to the caller.
	senderKeyPairInfo = keypairs.GlobalKeyPairStore.GetKeyPairInfo(useSenderName)
	if senderKeyPairInfo == nil {
		return nil, nil, fmt.Errorf("Unable to locate sender's keypair for name \"%s\"\n", useSenderName)
	}

	if strings.ToLower(senderKeyPairInfo.Name) == "default" {
		// now that we have retrieved the sender's key, we can set the sender name to
		// something else if needed for the bundle details
		profile := helpers.GlobalConfig.GetCurrentProfile()
		if profile != nil && profile.DefaultKeypairName != "" {
			senderKeyPairInfo.Name = profile.DefaultKeypairName
		} else {
			senderKeyPairInfo.Name = "Not Provided"
		}
	}

	receiverEntity := keystore.GlobalKeyStore.GetKey(localBundleCommandVals.toName)
	if receiverEntity == nil {
		return nil, nil, fmt.Errorf("receiver key not located for name \"%s\"", localBundleCommandVals.toName)
	}

	// The returned Entity and encapsulated keys are cloned during the GetKey() call, so ok to own them
	// here and just return them without cloning again.  Maybe a bit of an optimization and mem cost savings.
	return receiverEntity.PublicKeys, senderKeyPairInfo, nil
}

// getLocalKeysForBundleWrite will return a set of keys using the default read and write keypairs in the profile's keypair store
func getLocalKeysForBundleWrite() (receiverKeyInfo *security.KeyInfo, senderKeyPairInfo *security.KeyPairInfo, err error) {
	kpiKeypairStoreRead := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreReads)
	if kpiKeypairStoreRead == nil {
		return nil, nil, errors.New("store default read keypair not found")
	}

	receiverCipherPublicKey, receiverSigningPublicKey, err := kpiKeypairStoreRead.PublicKeys()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to obtain public keys from read keypair: %w", err)
	}

	receiverKeyInfo, err = security.NewKeyInfo(
		"local",
		receiverCipherPublicKey,
		receiverSigningPublicKey,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to build receiver key info: %w", err)
	}

	kpiKeypairStoreWrite := keypairs.GlobalKeyPairStore.GetKeyPairInfo(helpers.KeyPairNameForKeyStoreWrites)
	if kpiKeypairStoreWrite == nil {
		return nil, nil, errors.New("store default write keypair not found")
	}

	if err != nil {
		return nil, nil, fmt.Errorf("unable to build sender key info: %w", err)
	}

	// The kpiKeypairStoreWrite is the sender, functionally. Also, it is a returned clone from GetKeyPairInfo()
	// call, so ok to return without cloning here.
	return receiverKeyInfo, kpiKeypairStoreWrite, nil
}

func inferOutputTargetForBundle() (outputTargetWasInferred bool) {
	switch localBundleCommandVals.inputSource {
	case keystore.InputSourceConsole:
		localBundleCommandVals.outputTarget = keystore.OutputTargetConsole
		return true
	case keystore.InputSourceClipboard:
		localBundleCommandVals.outputTarget = keystore.OutputTargetClipboard
		return false
	case keystore.InputSourceFile:
		cwd, err := os.Getwd()
		if err != nil {
			logger.Errorfln("unable to determine the current working directory: %s", err)
			return false
		}

		localBundleCommandVals.outputTarget = keystore.OutputTargetPath
		localBundleCommandVals.outputPath = cwd

		return true
	case keystore.InputSourceDirs:
		logger.Errorfln("Input-source DIRS requires output target of type FILE")
		return false
	default:
		logger.Errorfln("Unsupported input source detected: %d\n", int(localBundleCommandVals.inputSource))
		return false
	}
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

func validateInputDirs() error {
	if localBundleCommandVals.inputDir != "" {
		isFound, isDir, err := helpers.FileExistsWithDetails(localBundleCommandVals.inputDir)
		if !isFound {
			return fmt.Errorf("input-dir does not exist: %s", localBundleCommandVals.inputDir)
		}

		if !isDir {
			return fmt.Errorf("input-dir is a file.  For input-source type DIRS, it must refer to a directory: %s", localBundleCommandVals.inputDir)
		}

		if err != nil {
			return fmt.Errorf("an error occured trying to obtain info for input-dir \"%s\": %s",
				localBundleCommandVals.inputDir, err)
		}

		return nil
	}

	if localBundleCommandVals.inputDescriptorPath == "" {
		return errors.New("no value supplied for input-dir or input-descriptor.  For input-source type DIRS, you must supply one of these")
	}

	if !helpers.FileExists(localBundleCommandVals.inputDescriptorPath) {
		return fmt.Errorf("input-descriptor not found: %s", localBundleCommandVals.inputDescriptorPath)
	}

	pathList, err := getDescriptorPaths(localBundleCommandVals.inputDescriptorPath)
	if err != nil {
		return fmt.Errorf("error reading descriptor: %s", err)
	}

	for _, path := range pathList {
		if path == "" {
			continue
		}

		isFound, isDir, err := helpers.FileExistsWithDetails(path)
		if !isFound {
			return fmt.Errorf("descriptor path reference does not exist: \"%s\"", path)
		}

		if !isDir {
			return fmt.Errorf("descriptor reference is a file.  All descriptor refs must refer to a directory: %s", path)
		}

		if err != nil {
			return fmt.Errorf("an error occured trying to obtain info for descriptor reference \"%s\": %s",
				path, err)
		}
	}

	return nil
}

func getDescriptorPaths(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var paths []string
	for scanner.Scan() {
		paths = append(paths, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

func validateOutputFile() error {
	// inferOutputTargetForBundle will have done some of this already
	if localBundleCommandVals.outputFile == "" {
		// we will attempt to infer the output file path from the input file path
		if localBundleCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output file path is provided and the input source is not \"file\", so unable to infer an output file path")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input file path is provided, so unable to infer an output file path")
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
		if localBundleCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output filename is provided and the input source is not \"file\", so unable to infer an output filename")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input filename is provided, so unable to infer an output filename")
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
		if localBundleCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output file path is provided and the input source is not \"file\", so unable to infer an output file path")
		}

		if localBundleCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input file path is provided, so unable to infer an output file path")
		}

		inputPath, _ := filepath.Split(localBundleCommandVals.inputFilePath)
		localBundleCommandVals.outputFile = inputPath

		return nil
	}

	// a path is defined, let's validate that it does exist, and it is a path
	exists, isDir := helpers.PathExistsInfo(localBundleCommandVals.outputPath)
	if !exists {
		return fmt.Errorf("provided output path does not exist: %s", localBundleCommandVals.outputPath)
	}

	if !isDir {
		return fmt.Errorf("provided output path references a file and not a path: %s", localBundleCommandVals.outputPath)
	}

	return nil
}

func getInputReader() (io.Reader, error) {
	switch localBundleCommandVals.inputSource {
	case keystore.InputSourceConsole:
		return getConsoleReader()
	case keystore.InputSourceClipboard:
		return getClipboardReader()
	case keystore.InputSourceFile:
		return getFileReader()
	case keystore.InputSourcePiped:
		return getPipedReader()
	case keystore.InputSourceDirs:
		return getDirsReader()
	}

	return nil, errors.New("unknown input source obtaining stream reader")
}

func getConsoleReader() (io.Reader, error) {
	inputLines, err := helpers.GetConsoleMultipleInputLines("bundle")
	if err != nil {
		return nil, fmt.Errorf("unable to get user input: %s", err)
	}

	if localBundleCommandVals.outputTarget == keystore.OutputTargetPath && localBundleCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputTarget = keystore.OutputTargetFile
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

	if localBundleCommandVals.outputTarget == keystore.OutputTargetPath && localBundleCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputTarget = keystore.OutputTargetFile
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

	if localBundleCommandVals.outputTarget == keystore.OutputTargetPath && localBundleCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localBundleCommandVals.outputTarget = keystore.OutputTargetFile
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

func getDirsReader() (io.Reader, error) {
	var err error
	localBundleSettings.mdsr, err = streams.NewMultiDirectoryStreamReader(streams.WithCompression())
	if err != nil {
		return nil, fmt.Errorf("unable to create multi directory input stream: %w", err)
	}

	localBundleSettings.cipherWriter.OutputBundleInfo.InputSource = cipherio.BundleInputSourceMultiDir

	if localBundleCommandVals.inputDir != "" {
		// Everything has already been validated, no need to do checks again, we'll assume all is good
		_, err := localBundleSettings.mdsr.AddDir(
			localBundleCommandVals.inputDir,
			streams.WithItemDetails(),
			streams.WithEmptyPaths())

		if err != nil {
			return nil, fmt.Errorf("unable to add directory \"%s\" to input stream: %w",
				localBundleCommandVals.inputDir,
				err)
		}
	} else {
		// Due to prior validations, we can assume the descriptor path is what we want to use now
		// Everything has already been validated, no need to do checks again, we'll assume all is good

		pathList, err := getDescriptorPaths(localBundleCommandVals.inputDescriptorPath)
		if err != nil {
			return nil, fmt.Errorf("unable to add descriptor paths: %w", err)
		}

		for _, path := range pathList {
			_, err := localBundleSettings.mdsr.AddDir(
				path,
				streams.WithItemDetails(),
				streams.WithEmptyPaths())

			if err != nil {
				return nil, fmt.Errorf("unable to add directory \"%s\" to input stream: %w",
					path,
					err)
			}
		}
	}

	streamReader, err := localBundleSettings.mdsr.StartStream()
	if err != nil {
		return nil, fmt.Errorf("unable to start multi directory stream: %w", err)
	}

	return streamReader, nil
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
