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

package helpers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

type InputResponseVal int

const (
	InputResponseValNull InputResponseVal = iota
	InputResponseValYes
	InputResponseValNo
)

const MAX_TRY_RECOUNTS = 5

func GetConsoleInputLine(promptText string) (inputLine string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error reading input: %s", r)
		}
	}()

	if promptText != "" {
		fmt.Printf("%s: ", promptText)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err = scanner.Err()
	if err != nil {
		return "", err
	}

	return scanner.Text(), nil
}

func GetConsoleRequiredInputLine(promptText, valueName string) (inputLine string, err error) {
	tryCount := 1
	tryCountText := ""
	for {
		if tryCount > 1 {
			tryCountText = fmt.Sprintf("(Try %d of %d)  ", tryCount, MAX_TRY_RECOUNTS)
		}
		inputLine, err = GetConsoleInputLine(tryCountText + promptText)
		if err != nil {
			return "", err
		}

		if inputLine != "" {
			return inputLine, err
		}

		tryCount += 1
		if tryCount > MAX_TRY_RECOUNTS {
			return "", fmt.Errorf("exceeded max tries of %d", MAX_TRY_RECOUNTS)
		}

		fmt.Printf("\"%s\" is required.  Please enter a value or CTRL-C to abort entry\n", valueName)
	}
}

func GetConsoleInputChar(allowedChars string) (inputChar string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error reading input: %s", r)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		var char rune
		char, _, err = reader.ReadRune()
		if err != nil {
			return "", err
		}

		inputChar = string(char)
		if strings.ContainsAny(inputChar, allowedChars) {
			return inputChar, err
		}
	}
}

// GetYesNoInput will display an optional message and append "Yes/No",
// then wait for user input.  If input is not valid, it will check again
func GetYesNoInput(inputMessage string, nullVal InputResponseVal) (InputResponseVal, error) {
	for {
		promptText := ""
		if inputMessage == "" {
			promptText = fmt.Sprintf("y/N: ")
		} else {
			promptText = fmt.Sprintf("%s (y/N) ", inputMessage)
		}

		text, err := GetConsoleInputLine(promptText)
		if err != nil {
			return InputResponseValNull, err
		}

		text = strings.Trim(text, " \n\t")
		if text == "" {
			return nullVal, nil
		}

		if CompareStrings(text, "y") == 0 {
			return InputResponseValYes, nil
		}

		if CompareStrings(text, "n") == 0 {
			return InputResponseValNo, nil
		}

		fmt.Println("Invalid response: expect \"Y\", \"y\", \"N\", \"n\" or empty for default value")
	}
}

// GetInputFromList will display a list of options and allow user to select one of them or cancel
type InputListItem struct {
	Option string
	Label  string
}

func GetInputFromList(inputMessage string, listItems []InputListItem, nullOption string) (string, error) {
	currentAttemptNumber := 0
	for {
		currentAttemptNumber += 1
		fmt.Println(inputMessage + " ...")
		for _, listItem := range listItems {
			fmt.Printf("  %s: %s\n", listItem.Option, listItem.Label)
		}

		fmt.Println("")
		promptText := fmt.Sprintf("Select (ENTER for %s)", nullOption)
		if currentAttemptNumber > 1 {
			promptText = fmt.Sprintf("Select (attempt %d of %d): ", currentAttemptNumber, MAX_TRY_RECOUNTS)
		}

		text, err := GetConsoleInputLine(promptText)
		if err != nil {
			return nullOption, err
		}

		text = strings.Trim(text, " \n\t")
		if text == "" {
			return nullOption, nil
		}

		for _, listItem := range listItems {
			if CompareStrings(listItem.Option, text) == 0 {
				return listItem.Option, nil
			}
		}

		fmt.Println("** Invalid response **")
		fmt.Println("")

		if currentAttemptNumber == MAX_TRY_RECOUNTS {
			return "", errors.New("retry count exceeded")
		}
	}
}

func GetPassword() ([]byte, error) {
	bytepw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println("")
	if err != nil {
		return nil, fmt.Errorf("failed entering password: %w", err)
	}

	return bytepw, nil
}

func GetPasswordWithConfirm(label string) ([]byte, error) {
	password, err := GetPassword()
	if err != nil {
		return nil, err
	}

	// Todo: do we need to confirm empty passwords?  We'll assume now for now
	if len(password) == 0 {
		return nil, nil
	}

	confirmed := false
	for j := 0; j < 5; j++ {
		if j > 0 {
			fmt.Printf(
				"Please re-enter the password to confirm it or CTRL-C to abort (attempt %d of %d): ",
				j+1,
				5,
			)
		} else {
			fmt.Printf("Please re-enter the password to confirm it or CTRL-C to abort: ")
		}

		passwordValidate, err := GetPassword()
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("error re-entering the password: %w", err)
		}

		if bytes.Compare(password, passwordValidate) == 0 {
			confirmed = true
			break
		}

		fmt.Println("** Re-entry does not match **")
	}

	fmt.Println()
	if !confirmed {
		return nil, errors.New("failed to re-enter password")
	}

	return password, nil
}

// AcquireKey will check to see if the key env var is available.  If not,
// it will prompt the user for the key
func AcquireKey(ProfileName string) ([]byte, error) {
	keyVal := os.Getenv("BB_" + strings.ToUpper(GetEnvSafeName(ProfileName)) + "_KEY")
	if keyVal != "" {
		return []byte(keyVal), nil
	}

	// No env var found, prompt the user for it
	fmt.Printf("Enter key for keypair store in profile \"%s\": ", ProfileName)
	keyBytes, err := GetPassword()
	if err != nil {
		return nil, fmt.Errorf("unable to get password from user input: %w", err)
	}

	return keyBytes, nil
}

func GetConsoleMultipleInputLines(labelText string) (inputLines []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error reading input: %s", r)
		}
	}()

	fmt.Printf("Enter one or more lines for the %s input. Enter an empty line to stop. CTRL-C to cancel.\n", labelText)

	for {
		lineText, err := GetConsoleInputLine(" ")
		if err != nil {
			fmt.Println("")
			return nil, err
		}

		if lineText == "" {
			fmt.Println("")
			return inputLines, nil
		}

		inputLines = append(inputLines, lineText)
	}
}
