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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keystore"
)

// removeKeyCmd represents the key command
var removeKeyCmd = &cobra.Command{
	Use:   "key [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Will remove the referenced key from the keystore",
	Long:  "Will remove the referenced key from the keystore",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		var keyName string
		switch len(args) {
		case 0:
			_ = cmd.Help()
			return
		case 1:
			keyName = args[0]
		}

		removeKey(keyName)
	},
}

func init() {
	removeCmd.AddCommand(removeKeyCmd)
}

func removeKey(keyName string) {
	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to remove key: keystore not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	entity := keystore.GlobalKeyStore.GetKey(keyName)
	if entity == nil {
		fmt.Printf("No key was found with name \"%s\"\n", keyName)
		return
	}

	fmt.Println("Entity info located...")
	entity.Print()
	fmt.Println("")
	response, err := helpers.GetYesNoInput(fmt.Sprintf("Are you sure you wish to remove the key \"%s\"?", keyName), helpers.InputResponseValNo)
	if err != nil {
		fmt.Printf("Unable to confirm removal of key: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if response != helpers.InputResponseValYes {
		fmt.Println("User aborted removal request")
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	found, err := keystore.GlobalKeyStore.RemoveEntity(keyName)
	if !found {
		fmt.Printf("Was unable to located the key during removal: \"%s\"\n", keyName)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if err != nil {
		fmt.Printf("Unable to remove key named \"%s\": %s\n", keyName, err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = keystore.GlobalKeyStore.WriteToFile("")
	if err != nil {
		fmt.Printf("Unable to save keystore file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Key removed.")
}
