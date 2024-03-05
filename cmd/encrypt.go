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
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keystore"
	"github.com/thoughtrealm/bumblebee/logger"
	"github.com/thoughtrealm/bumblebee/security"
	"github.com/thoughtrealm/bumblebee/streams"
	"github.com/thoughtrealm/bumblebee/symfiles"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypts a file or input using symmetric encryption and a user supplied key",
	Long:  "Encrypts a file or input using symmetric encryption and a user supplied key",
	Run: func(cmd *cobra.Command, args []string) {
		encryptData()
	},
}

type encryptCommandVals struct {
	// The user supplied key to encrypt the input with
	symmetricKey []byte

	// Commande line provided symmetric key
	symmetricKeyInputText string

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
}

var localEncryptCommandVals = &encryptCommandVals{}

type encryptSettings struct {
	totalBytesWritten  int
	mdsr               streams.StreamReader
	symFilePayloadType symfiles.SymFilePayload
}

var localEncryptSettings = &encryptSettings{}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.inputSourceText, "input-source", "i", "console", "The type of the input source.  Should be one of: console, clipboard, file or dirs.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.inputFilePath, "input-file", "f", "", "The name of a file to use for input. Only relevant if input-source is file.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.inputDir, "input-dir", "", "", "The name of a directory to use for input. Only relevant if input-source is dirs.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.inputDescriptorPath, "input-descriptor", "", "", "The name of a file that contains a list of directories for input. Only relevant if input-source is dirs.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.outputTargetText, "output-target", "o", "", "The output target.  Should be one of: console, clipboard or file.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.outputFile, "output-file", "y", "", "The file name to use for output. Only relevant if output-target is FILE.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.outputPath, "output-path", "p", "", "The path name to use for output. Only relevant if output-target is PATH.")
	encryptCmd.Flags().StringVarP(&localEncryptCommandVals.symmetricKeyInputText, "key", "", "", "The key to use for encrypted the data. Prompted for if not provided. Prompt entry is recommended.")
}

func encryptData() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered in encryptData(): %s\n", r)
		}
	}()

	var err error

	if localEncryptCommandVals.inputSourceText == "" {
		if localEncryptCommandVals.inputFilePath != "" {
			localEncryptCommandVals.inputSourceText = "file"
		} else if localEncryptCommandVals.inputDir != "" {
			localEncryptCommandVals.inputSourceText = "dirs"
		} else if localEncryptCommandVals.inputDescriptorPath != "" {
			localEncryptCommandVals.inputSourceText = "dirs"
		}
	}

	if localEncryptCommandVals.outputTargetText == "" && localEncryptCommandVals.outputFile != "" {
		localEncryptCommandVals.outputTargetText = "file"
	}

	if localEncryptCommandVals.outputTargetText == "" && localEncryptCommandVals.outputPath != "" {
		localEncryptCommandVals.outputTargetText = "path"
	}

	// do this check after the other inference checks above relating to no supplied value for inputSourceText
	if localEncryptCommandVals.inputSourceText == "" && helpers.CheckIsPiped() {
		localEncryptCommandVals.inputSourceText = "piped"
	}

	if localEncryptCommandVals.inputSourceText == "" {
		localEncryptCommandVals.inputSourceText = "console"
	}
	if localEncryptCommandVals.outputTargetText == "" {
		localEncryptCommandVals.outputTargetText = "console"
	}

	localEncryptCommandVals.inputSource = keystore.TextToInputSource(localEncryptCommandVals.inputSourceText)
	if localEncryptCommandVals.inputSource == keystore.InputSourceUnknown {
		fmt.Println("No input-source or invalid input-source provided.  --input-source is required.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localEncryptCommandVals.inputSource == keystore.InputSourceConsole && localEncryptCommandVals.outputTargetText == "" {
		localEncryptCommandVals.outputTargetText = "console"
	}

	localEncryptCommandVals.outputTarget = keystore.TextToOutputTarget(localEncryptCommandVals.outputTargetText)
	if localEncryptCommandVals.outputTarget == keystore.OutputTargetUnknown {
		fmt.Println("No output-target or invalid output-target provided.  --output-target is required.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localEncryptCommandVals.inputSource == keystore.InputSourceFile {
		err = validateInputFileForEncrypt()
		if err != nil {
			fmt.Printf("Input file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localEncryptCommandVals.inputSource == keystore.InputSourceDirs {
		err = validateInputDirsForEncrypt()
		if err != nil {
			fmt.Printf("Input dirs invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localEncryptCommandVals.outputTargetText == "" {
		if !inferOutputTargetForEncrypt() {
			fmt.Println("Unable to infer output-target based on input-source.  You must provide a value for --output-target.")
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else {
		localEncryptCommandVals.outputTarget = keystore.TextToOutputTarget(localEncryptCommandVals.outputTargetText)
		if localEncryptCommandVals.outputTarget == keystore.OutputTargetUnknown {
			fmt.Printf("Unknown output-target: \"%s\"\n", localEncryptCommandVals.outputTargetText)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localEncryptCommandVals.inputSource == keystore.InputSourceDirs &&
		localEncryptCommandVals.outputTarget != keystore.OutputTargetFile {
		fmt.Println("Incorrect output target for input source DIRS.  Output target MUST BE of type FILE.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if localEncryptCommandVals.outputTarget == keystore.OutputTargetFile {
		err = validateOutputFileForEncrypt()
		if err != nil {
			fmt.Printf("Output file invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localEncryptCommandVals.outputTarget == keystore.OutputTargetPath {
		err = validateOutputPathForEncrypt()
		if err != nil {
			fmt.Printf("Output path invalid: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if localEncryptCommandVals.symmetricKeyInputText != "" {
		localEncryptCommandVals.symmetricKey = []byte(localEncryptCommandVals.symmetricKeyInputText)
	} else {
		err = getEncryptKey()
		if err != nil {
			fmt.Printf("Unable to acquire data key: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if len(localEncryptCommandVals.symmetricKey) == 0 {
		// This can't really happen, but check in case anyway
		fmt.Println("No key provided")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	defer security.Wipe(localEncryptCommandVals.symmetricKey)

	inputReader, err := getInputReaderForEncrypt()
	if err != nil {
		fmt.Printf("Unable to acquire an input reader: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	outputWriter, err := getOutputWriterForEncrypt()
	if err != nil {
		fmt.Printf("Unable to acquire an output writer: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	symFileWriter, err := symfiles.NewSymFileWriter(localEncryptCommandVals.symmetricKey)
	if err != nil {
		fmt.Printf("Unable to initialize symFile instance: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	bytesWritten, err := symFileWriter.WriteSymFileToWriterFromReader(
		inputReader, outputWriter, localEncryptSettings.symFilePayloadType)

	if err != nil {
		fmt.Printf("Error encrypting output: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Encrypt complete.")
	p := message.NewPrinter(language.English)
	_, _ = p.Printf("Bytes Written: %d", bytesWritten)
}

func validateInputFileForEncrypt() error {
	if localEncryptCommandVals.inputFilePath == "" {
		return errors.New("no input file provided")
	}

	if !helpers.FileExists(localEncryptCommandVals.inputFilePath) {
		return fmt.Errorf("input file path does not exist: %s", localEncryptCommandVals.inputFilePath)
	}

	return nil
}

func validateInputDirsForEncrypt() error {
	if localEncryptCommandVals.inputDir != "" {
		isFound, isDir := helpers.PathExistsInfo(localEncryptCommandVals.inputDir)
		if !isFound {
			return fmt.Errorf("input-dir does not exist: %s", localEncryptCommandVals.inputDir)
		}

		if !isDir {
			return fmt.Errorf("input-dir is a file.  For input-source type DIRS, it must refer to a directory: %s", localEncryptCommandVals.inputDir)
		}

		return nil
	}

	if localEncryptCommandVals.inputDescriptorPath == "" {
		return errors.New("no value supplied for input-dir or input-descriptor.  For input-source type DIRS, you must supply one of these")
	}

	if !helpers.FileExists(localEncryptCommandVals.inputDescriptorPath) {
		return fmt.Errorf("input-descriptor not found: %s", localEncryptCommandVals.inputDescriptorPath)
	}

	pathList, err := getDescriptorPaths(localEncryptCommandVals.inputDescriptorPath)
	if err != nil {
		return fmt.Errorf("error reading descriptor: %s", err)
	}

	for _, path := range pathList {
		if path == "" {
			continue
		}

		isFound, isDir := helpers.PathExistsInfo(path)
		if !isFound {
			return fmt.Errorf("descriptor path reference does not exist: \"%s\"", path)
		}

		if !isDir {
			return fmt.Errorf("descriptor reference is a file.  All descriptor refs must refer to a directory: %s", path)
		}
	}

	return nil
}

func inferOutputTargetForEncrypt() bool {
	switch localEncryptCommandVals.inputSource {
	case keystore.InputSourceConsole:
		localEncryptCommandVals.outputTarget = keystore.OutputTargetConsole
		return true
	case keystore.InputSourceClipboard:
		localEncryptCommandVals.outputTarget = keystore.OutputTargetClipboard
		// Todo: Bundle returns false for this, but we are going to try true and see how it works
		return true
	case keystore.InputSourceFile:
		cwd, err := os.Getwd()
		if err != nil {
			logger.Errorfln("unable to determine the current working directory: %s", err)
			return false
		}

		localEncryptCommandVals.outputTarget = keystore.OutputTargetPath
		localEncryptCommandVals.outputPath = cwd

		return true
	case keystore.InputSourceDirs:
		logger.Errorfln("Input-source DIRS requires output target of type FILE")
		return false
	default:
		logger.Errorfln("Unsupported input source detected: %d\n", int(localEncryptCommandVals.inputSource))
		return false
	}
}

func validateOutputFileForEncrypt() error {
	// inferOutputTargetForEncrypt will have done some of this already
	if localEncryptCommandVals.outputFile == "" {
		// we will attempt to infer the output file path from the input file path
		if localEncryptCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output file path is provided and the input source is not \"file\". Unable to infer an output file path")
		}

		if localEncryptCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input file path is provided. Unable to infer an output file path")
		}

		// an output filename with an extension of "bsym"
		localEncryptCommandVals.outputFile = filepath.Join(
			localEncryptCommandVals.inputFilePath, helpers.ReplaceFileExt(
				localEncryptCommandVals.inputFilePath,
				".bsym"))
	}

	// let's confirm that at least a filename has been provided at this point
	outputFilePath, outputFileName := filepath.Split(localEncryptCommandVals.outputFile)
	if outputFileName == "" {
		// we will attempt to infer the output filename from the input file path
		if localEncryptCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output filename is provided and the input source is not \"file\". Unable to infer an output filename")
		}

		if localEncryptCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input filename is provided. Unable to infer an output filename")
		}

		_, inputFileName := filepath.Split(localEncryptCommandVals.inputFilePath)
		outputFileName = helpers.ReplaceFileExt(inputFileName, ".bsym")
		localEncryptCommandVals.outputFile = filepath.Join(outputFilePath, outputFileName)
	}

	// We will assume that whatever filepath has been provided or inferred is ok.  If the path doesn't exist,
	// or something else is invalid about the path, the OS will fail with an error during create and we
	// will return that to the user at that point.  We won't implement a bunch of code here to
	// reproduce that os validation behavior.

	return nil
}

// validateOutputPathForEncrypt will do two things...
//   - If there is no path defined yet, it will take the path from the input path.
//   - If there is a path defined, it will throw an error if the path has a file name defined, since we only want a path
func validateOutputPathForEncrypt() error {
	if localEncryptCommandVals.outputPath == "" {
		// we will attempt to infer the output file path from the input file path
		if localEncryptCommandVals.inputSource != keystore.InputSourceFile {
			return errors.New("output target set to \"file\", however no output file path is provided and the input source is not \"file\". Unable to infer an output file path")
		}

		if localEncryptCommandVals.inputFilePath == "" {
			return errors.New("output target set to \"file\", however no output file path is provided and no input file path is provided. Unable to infer an output file path")
		}

		inputPath, _ := filepath.Split(localEncryptCommandVals.inputFilePath)
		localEncryptCommandVals.outputFile = inputPath

		return nil
	}

	// a path is defined, let's validate that it does exist, and it is a path
	exists, isDir := helpers.PathExistsInfo(localEncryptCommandVals.outputPath)
	if !exists {
		return fmt.Errorf("provided output path does not exist: %s", localEncryptCommandVals.outputPath)
	}

	if !isDir {
		return fmt.Errorf("provided output path references a file and not a path: %s", localEncryptCommandVals.outputPath)
	}

	return nil
}

func getEncryptKey() error {
	fmt.Printf("\nEnter a password/key for the encryted data: ")
	key, err := helpers.GetPasswordWithConfirm("Enter password for encrypted data")
	if err != nil {
		return fmt.Errorf("Unable to acquire key for data: %w", err)
	}

	localEncryptCommandVals.symmetricKey = bytes.Clone(key)
	security.Wipe(key)

	return nil
}

func getInputReaderForEncrypt() (io.Reader, error) {
	switch localEncryptCommandVals.inputSource {
	case keystore.InputSourceConsole:
		return getConsoleReaderForEncrypt()
	case keystore.InputSourceClipboard:
		return getClipboardReaderForEncrypt()
	case keystore.InputSourceFile:
		return getFileReaderForEncrypt()
	case keystore.InputSourcePiped:
		return getPipedReaderForEncrypt()
	case keystore.InputSourceDirs:
		return getDirsReaderForEncrypt()
	}

	return nil, errors.New("unknown input source obtaining stream reader")
}

func getConsoleReaderForEncrypt() (io.Reader, error) {
	inputLines, err := helpers.GetConsoleMultipleInputLines("encrypt")
	if err != nil {
		return nil, fmt.Errorf("unable to get user input: %s", err)
	}

	if localEncryptCommandVals.outputTarget == keystore.OutputTargetPath && localEncryptCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localEncryptCommandVals.outputTarget = keystore.OutputTargetFile
		localEncryptCommandVals.outputFile = filepath.Join(localEncryptCommandVals.outputPath, "bee.console.ext")
	}

	localEncryptSettings.symFilePayloadType = symfiles.SymFilePayloadDataStream
	inputBytes := []byte(strings.Join(inputLines, "\n"))
	inputBuff := bytes.NewBuffer(inputBytes)
	return inputBuff, nil
}

func getClipboardReaderForEncrypt() (io.Reader, error) {
	data, err := helpers.ReadFromClipboard()
	if err != nil {
		return nil, fmt.Errorf("unable to read from clipboard: %w", err)
	}

	reader, err := helpers.NewTextScanner(data)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize text scanner from clipboard input: %s", err)
	}

	if localEncryptCommandVals.outputTarget == keystore.OutputTargetPath && localEncryptCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localEncryptCommandVals.outputTarget = keystore.OutputTargetFile
		localEncryptCommandVals.outputFile = filepath.Join(localEncryptCommandVals.outputPath, "bee.clipboard.ext")
	}

	localEncryptSettings.symFilePayloadType = symfiles.SymFilePayloadDataStream
	return reader, nil
}

func getPipedReaderForEncrypt() (io.Reader, error) {
	pipeBuffer := bytes.NewBuffer(nil)
	_, err := pipeBuffer.ReadFrom(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to read piped input from stdin: %s", err)
	}

	reader, err := helpers.NewTextScanner(pipeBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to initialize pipe text scanner from pipe input: %s", err)
	}

	if localEncryptCommandVals.outputTarget == keystore.OutputTargetPath && localEncryptCommandVals.inputFilePath == "" {
		// if an output target of PATH is specified, we need to add a file name if one is not specified via the inputFilePath
		// writeToPath
		localEncryptCommandVals.outputTarget = keystore.OutputTargetFile
		localEncryptCommandVals.outputFile = filepath.Join(localBundleCommandVals.outputPath, "bee.piped.ext")
	}

	localEncryptSettings.symFilePayloadType = symfiles.SymFilePayloadDataStream
	return reader, nil
}

func getFileReaderForEncrypt() (io.Reader, error) {
	file, err := os.Open(localEncryptCommandVals.inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to initiate input file stream: %w", err)
	}

	localEncryptSettings.symFilePayloadType = symfiles.SymFilePayloadDataStream
	return file, nil
}

func getDirsReaderForEncrypt() (io.Reader, error) {
	var err error
	localEncryptSettings.mdsr, err = streams.NewMultiDirectoryStreamReader(streams.WithCompression())
	if err != nil {
		return nil, fmt.Errorf("unable to create multi directory input stream: %w", err)
	}

	localEncryptSettings.symFilePayloadType = symfiles.SymFilePayloadDataMultiDir

	if localEncryptCommandVals.inputDir != "" {
		// Everything has already been validated, no need to do checks again, we'll assume all is good
		_, err := localEncryptSettings.mdsr.AddDir(
			localEncryptCommandVals.inputDir,
			streams.WithItemDetails(),
			streams.WithEmptyPaths())

		if err != nil {
			return nil, fmt.Errorf("unable to add directory \"%s\" to input stream: %w",
				localEncryptCommandVals.inputDir,
				err)
		}
	} else {
		// Due to prior validations, we can assume the descriptor path is what we want to use now
		// Everything has already been validated, no need to do checks again, we'll assume all is good

		pathList, err := getDescriptorPaths(localEncryptCommandVals.inputDescriptorPath)
		if err != nil {
			return nil, fmt.Errorf("unable to add descriptor paths: %w", err)
		}

		for _, path := range pathList {
			_, err := localEncryptSettings.mdsr.AddDir(
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

	streamReader, err := localEncryptSettings.mdsr.StartStream()
	if err != nil {
		return nil, fmt.Errorf("unable to start multi directory stream: %w", err)
	}

	return streamReader, nil
}

func getOutputWriterForEncrypt() (w io.Writer, err error) {
	return nil, errors.New("getOutputWriterForEncrypt not implemented")
}
