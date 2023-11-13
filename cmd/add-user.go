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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thoughtrealm/bumblebee/keystore"
)

type addUserSubcommandVals struct {
	cipherPublicKey  string
	signingPublicKey string
}

var localAddUserSubcommandVals = &addUserSubcommandVals{}

// addUserCmd represents the key sub-command for the add command.
// We maintain the public keys for external users.
var addUserCmd = &cobra.Command{
	Use:   "user <keyName> --cipher <key> --signing <key>",
	Args:  cobra.ExactArgs(1),
	Short: "Adds a new user using the provided cipher and signing public keys",
	Long:  "Adds a new user using the provided cipher and signing public keys",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Exact args are set by Args property above. Only need to check for cipher and signing keys
		if localAddUserSubcommandVals.cipherPublicKey == "" {
			fmt.Println("No cipher public key provided.  Cipher public key is required.")
			return
		}

		if localAddUserSubcommandVals.signingPublicKey == "" {
			fmt.Println("No signing public key provided.  signing public key is required.")
			return
		}

		addNewKey(args[0])
	},
}

func init() {
	addCmd.AddCommand(addUserCmd)
	addUserCmd.Flags().StringVarP(&localAddUserSubcommandVals.cipherPublicKey, "cipher", "c", "", "The value for the public cipher key")
	addUserCmd.Flags().StringVarP(&localAddUserSubcommandVals.signingPublicKey, "signing", "s", "", "The value for the public signing key")
}

func addNewKey(userName string) {
	var err error

	err = keystore.GlobalKeyStore.AddKey(
		userName,
		localAddUserSubcommandVals.cipherPublicKey,
		localAddUserSubcommandVals.signingPublicKey,
	)
	if err != nil {
		fmt.Printf("Unable to add new user: %v\n", errors.Unwrap(err))
	}

	fmt.Println("New user stored to file")
}
