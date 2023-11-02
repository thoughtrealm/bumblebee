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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/logger"
	"os"
	"strings"
)

type readOutputEncoding int

const (
	ReadOutputEncodingUnknown readOutputEncoding = iota
	ReadOutputEncodingBase64
	ReadOutputEncodingHex
	ReadOutputEncodingBlob
)

func textToReadOutputEncoding(val string) (readOutputEncoding, error) {
	switch strings.ToLower(strings.Trim(val, " \r\n\t")) {
	case "base64":
		return ReadOutputEncodingBase64, nil
	case "hex":
		return ReadOutputEncodingHex, nil
	case "blob":
		return ReadOutputEncodingBlob, nil
	default:
		return ReadOutputEncodingUnknown, fmt.Errorf("unknown readOutputEncoding name: %s", val)
	}
}

type readCommandVals struct {
	outputEncodingText string
	outputEncoding     readOutputEncoding
	inputFilePath      string
}

var localReadCommandVals = &readCommandVals{}

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Args:  cobra.MaximumNArgs(2),
	Short: "Reads a file and streams output for piping binary file inputs. Primarily for Windows, but should work on any platform.",
	Long:  "Reads a file and streams output for piping binary file inputs. Primarily for Windows, but should work on any platform.",
	Run: func(cmd *cobra.Command, args []string) {
		if localReadCommandVals.inputFilePath == "" && localReadCommandVals.outputEncodingText == "" {
			_ = cmd.Help()
			return
		}

		var err error
		localReadCommandVals.outputEncoding, err = textToReadOutputEncoding(localReadCommandVals.outputEncodingText)
		if err != nil {
			// Error is already formatted
			fmt.Printf("%s\n", err)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}

		if localReadCommandVals.inputFilePath == "" {
			fmt.Println("\"input-file\" is empty. \"input-file\" is required")
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}

		if !helpers.FileExists(localReadCommandVals.inputFilePath) {
			fmt.Printf("Value provided for input file path does not exist: \"%s\"\n", localReadCommandVals.inputFilePath)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}

		readInputFileToOutputStream()
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Flags().StringVarP(&localReadCommandVals.inputFilePath, "input-file", "f", "", "The path of the file to read")
	readCmd.Flags().StringVarP(&localReadCommandVals.outputEncodingText, "output-encoding", "o", "hex", "The encoding to output. Should be \"hex\" or \"base64\".")
}

func readInputFileToOutputStream() {
	fileIn, err := os.Open(localReadCommandVals.inputFilePath)
	if err != nil {
		fmt.Printf("Unable to open input file: \"%s\"\n", localReadCommandVals.inputFilePath)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	defer func() {
		_ = fileIn.Close()
	}()

	fileBuffer := bytes.NewBuffer(nil)
	bytesRead, err := fileBuffer.ReadFrom(fileIn)
	if err != nil {
		fmt.Printf("Error reading from input file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fileBytes := fileBuffer.Bytes()
	if len(fileBytes) == 0 {
		// Todo: Should this be an error or no?  We'll write to debug out in case it should not be an error.
		logger.Debugf("reading input-file returned no data: %s", localReadCommandVals.inputFilePath)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	// we use int64 here for consistency with buffer's output from ReadFrom().
	// this should need to be a large amount of data for this read.
	var bytesWritten int64
	switch localReadCommandVals.outputEncoding {
	case ReadOutputEncodingBase64:
		bytesWritten, err = writeBase64Output(fileBytes)
	case ReadOutputEncodingHex:
		bytesWritten, err = writeHexOutput(fileBytes)
	case ReadOutputEncodingBlob:
		bytesWritten, err = writeBlobOutput(fileBytes)
	default:
		// Should never happen, but if we should add a new type, this would tell us during testing
		// that we forgot to add support in this switch.
		fmt.Println("Unknown output encoding type on write")
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if err != nil {
		fmt.Printf("Unable to write output: %s", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	logger.Debugf("Stream output complete: %d bytes read, %d bytes written.", bytesRead, bytesWritten)
}

func writeHexOutput(fileBytes []byte) (totalBytesWritten int64, err error) {
	// Todo: For hex output, we will use 30 bytes per line for now... maybe offer line width as a flag at some point
	const BytesPerLine = 30

	var (
		bytesWritten int
	)

	bytesWritten, err = fmt.Fprintln(os.Stdout, ":start  :raw   :hex ========================================")
	if err != nil {
		return int64(bytesWritten), fmt.Errorf("error writing output format signature: %s", err)
	}

	totalBytesWritten += int64(bytesWritten)

	defer func() {
		if err != nil {
			// if we are exitting after an error, no need to send the footer line
			return
		}

		bytesWritten, err = fmt.Fprintln(os.Stdout, ":end =======================================================")
		totalBytesWritten += int64(bytesWritten)
		if err != nil {
			err = fmt.Errorf("error writing output format signature: %s", err)
			return
		}
	}()

	currentPos := 0
	for {
		if currentPos+BytesPerLine >= len(fileBytes) {
			bytesWritten, err = fmt.Fprintf(os.Stdout, "%x\n", fileBytes[currentPos:])
			return totalBytesWritten + int64(bytesWritten), err
		}

		bytesWritten, err = fmt.Fprintf(os.Stdout, "%x\n", fileBytes[currentPos:currentPos+BytesPerLine])
		totalBytesWritten += int64(bytesWritten)
		if err != nil {
			return totalBytesWritten, err
		}

		currentPos += BytesPerLine
		if currentPos >= len(fileBytes) {
			return totalBytesWritten, nil
		}
	}
}

func writeBase64Output(fileBytes []byte) (totalBytesWritten int64, err error) {
	// Todo: For base64 output, we will use 60 bytes per line for now... maybe offer line width as a flag at some point
	const BytesPerLine = 60

	var (
		bytesWritten int
		outputBytes  []byte
		sourceBytes  []byte
	)

	bytesWritten, err = fmt.Fprintln(os.Stdout, ":start  :raw   :base64 ====================================================")
	if err != nil {
		return int64(bytesWritten), fmt.Errorf("error writing output format signature: %s", err)
	}

	totalBytesWritten += int64(bytesWritten)

	defer func() {
		if err != nil {
			// if we are exitting after an error, no need to send the footer line
			return
		}

		bytesWritten, err = fmt.Fprintln(os.Stdout, ":end ======================================================================")
		totalBytesWritten += int64(bytesWritten)
		if err != nil {
			err = fmt.Errorf("error writing output format signature: %s", err)
			return
		}
	}()

	currentPos := 0
	for {
		if currentPos+BytesPerLine >= len(fileBytes) {
			sourceBytes = fileBytes[currentPos:]
			outputBytes = make([]byte, base64.RawStdEncoding.EncodedLen(len(sourceBytes)))
			base64.RawStdEncoding.Encode(outputBytes, sourceBytes)

			bytesWritten, err = fmt.Fprintf(os.Stdout, "%s\n", string(outputBytes))
			return totalBytesWritten + int64(bytesWritten), err
		}

		sourceBytes = fileBytes[currentPos : currentPos+BytesPerLine]
		outputBytes = make([]byte, base64.RawStdEncoding.EncodedLen(len(sourceBytes)))
		base64.RawStdEncoding.Encode(outputBytes, sourceBytes)

		bytesWritten, err = fmt.Fprintf(os.Stdout, "%s\n", string(outputBytes))
		totalBytesWritten += int64(bytesWritten)
		if err != nil {
			return totalBytesWritten, err
		}

		currentPos += BytesPerLine
		if currentPos >= len(fileBytes) {
			return totalBytesWritten, nil
		}
	}
}

func writeBlobOutput(fileBytes []byte) (totalBytesWritten int64, err error) {
	var bytesWritten int
	bytesWritten, err = os.Stdout.Write(fileBytes)
	totalBytesWritten += int64(bytesWritten)
	return totalBytesWritten, err
}
