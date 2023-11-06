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
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keystore"
)

// addKeyCmd represents the key sub-command for the add command.
// Currently, we only support adding public keys for external accounts, no private seeds for now
var addKeyCmd = &cobra.Command{
	Use:   "key <keyName> [publicKey]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Adds a new public key for external accounts",
	Long:  "Adds a new public key for external accounts",
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

		addNewKey(keyName, publicKey)
	},
}

func init() {
	addCmd.AddCommand(addKeyCmd)
}

func addNewKey(keyName, publicKey string) {
	var err error
	if keyName == "" {
		keyName, err = helpers.GetConsoleRequiredInputLine(
			"Enter name for the new key",
			"Name",
		)

		if err != nil {
			fmt.Printf("Unable to get key name: %s\n", err)
			return
		}
	}

	if publicKey == "" {
		publicKey, err = helpers.GetConsoleRequiredInputLine(
			"Enter the public key",
			"Public Key",
		)

		if err != nil {
			fmt.Printf("Unable to get Public Key: %s\n", err)
			return
		}
	}

	err = keystore.GlobalKeyStore.AddKey(keyName, []byte(publicKey))
	if err != nil {
		fmt.Printf("Unable to add new key: %v\n", errors.Unwrap(err))
	}

	fmt.Println("New key stored to file")
}
