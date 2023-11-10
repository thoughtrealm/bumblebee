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

type updateUserKeySubcommandVals struct {
	cipherPublicKey  string
	signingPublicKey string
}

var localUpdateUserKeySubcommandVals = &updateUserKeySubcommandVals{}

// updateUserKeyCmd represents the key command
var updateUserKeyCmd = &cobra.Command{
	Use:   "user-key <name> --cipher=cipherKey --signing=signingKey",
	Args:  cobra.ExactArgs(1),
	Short: "Will replace the public keys for a keystore key",
	Long:  "Will replace the public keys for a keystore key",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Max args is set by Args property above. Only need to check for 2 or less args
		var keyName string

		if localUpdateUserKeySubcommandVals.cipherPublicKey == "" &&
			localUpdateUserKeySubcommandVals.signingPublicKey == "" {

			fmt.Println("No keys provided.  Expected at least one of \"--cipher\" or \"signing\"", err)
			return
		}

		keyName = args[0]

		updateUserPublicKeys(keyName)
	},
}

func init() {
	updateCmd.AddCommand(updateUserKeyCmd)
	updateUserKeyCmd.Flags().StringVarP(&localUpdateUserKeySubcommandVals.cipherPublicKey, "cipher", "c", "", "The value for the public cipher key")
	updateUserKeyCmd.Flags().StringVarP(&localUpdateUserKeySubcommandVals.signingPublicKey, "signing", "s", "", "The value for the public signing key")
}

func updateUserPublicKeys(keyName string) {
	var err error
	var found bool

	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to update key: keystore not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	// the user may submit both keys, or just the cipher key, or just the signing key
	if localUpdateUserKeySubcommandVals.cipherPublicKey != "" &&
		localUpdateUserKeySubcommandVals.signingPublicKey != "" {
		// user provided both keys
		found, err = keystore.GlobalKeyStore.UpdatePublicKeys(
			keyName,
			localUpdateUserKeySubcommandVals.cipherPublicKey,
			localUpdateUserKeySubcommandVals.signingPublicKey)
		if !found {
			fmt.Printf("Unable to update keys: key not found with name \"%s\"\n", keyName)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else if localUpdateUserKeySubcommandVals.cipherPublicKey != "" {
		// user provided only the cipher key
		found, err = keystore.GlobalKeyStore.UpdateCipherPublicKey(
			keyName,
			localUpdateUserKeySubcommandVals.cipherPublicKey)
		if !found {
			fmt.Printf("Unable to update cipher key: key not found with name \"%s\"\n", keyName)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	} else {
		// user provided only the signing key
		found, err = keystore.GlobalKeyStore.UpdateSigningPublicKey(
			keyName,
			localUpdateUserKeySubcommandVals.signingPublicKey)
		if !found {
			fmt.Printf("Unable to update signing key: key not found with name \"%s\"\n", keyName)
			helpers.ExitCode = helpers.ExitCodeInvalidInput
			return
		}
	}

	if err != nil {
		fmt.Printf("Unable to update key(s): %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	// Save the changed key store
	err = keystore.GlobalKeyStore.WriteToFile("")
	if err != nil {
		fmt.Printf("Unable to update key: keystore could not update the file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Key updated and keystore file changes committed.")
}
