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

// updateKeyCmd represents the key command
var updateKeyCmd = &cobra.Command{
	Use:   "key <name> <publicKey>",
	Args:  cobra.MaximumNArgs(2),
	Short: "Will replace the public key for a keystore key",
	Long:  "Will replace the public key for a keystore key",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Max args is set by Args property above. Only need to check for 2 or less args
		var keyName string
		var publicKey string

		switch len(args) {
		case 0:
			_ = cmd.Help()
		case 1:
			keyName = args[0]
		case 2:
			keyName = args[0]
			publicKey = args[1]
		}

		updateKey(keyName, publicKey)
	},
}

func init() {
	updateCmd.AddCommand(updateKeyCmd)
}

func updateKey(keyName, publicKey string) {
	var err error
	if keyName == "" {
		keyName, err = helpers.GetConsoleRequiredInputLine(
			"Enter the name of the key to update",
			"Name",
		)

		if err != nil {
			fmt.Printf("Unable to get key name: %s\n", err)
			return
		}
	}

	if publicKey == "" {
		publicKey, err = helpers.GetConsoleRequiredInputLine(
			"Enter the new public key",
			"Public Key",
		)

		if err != nil {
			fmt.Printf("Unable to get Public Key: %s\n", err)
			helpers.ExitCode = helpers.ExitCodeInputError
			return
		}

		if publicKey == "" {
			fmt.Println("Nothing entred for public key.  Public key is required.")
			helpers.ExitCode = helpers.ExitCodeInputError
			return
		}
	}

	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to update key: keystore not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	found, err := keystore.GlobalKeyStore.UpdatePublicKey(keyName, publicKey)
	if !found {
		fmt.Printf("Unable to update key: key not found with name \"%s\"\n", keyName)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	// Save the change key store
	err = keystore.GlobalKeyStore.WriteToFile("")
	if err != nil {
		fmt.Printf("Unable to update key: keystore could not update the file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Key updated and keystore file changes committed.")
}
