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

// renameKeyCmd represents the key command
var renameKeyCmd = &cobra.Command{
	Use:   "key <oldName> <newName>",
	Args:  cobra.ExactArgs(2),
	Short: "Will change the  name of a key from oldName to newName",
	Long:  "Will change the  name of a key from oldName to newName",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Exact args is set by Args property above. So we should ALWAYS have 2 args
		oldKeyName := args[0]
		newKeyName := args[1]

		renameKey(oldKeyName, newKeyName)
	},
}

func init() {
	renameCmd.AddCommand(renameKeyCmd)
}

func renameKey(oldKeyName, newKeyName string) {
	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to rename key: keystore not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	found, err := keystore.GlobalKeyStore.RenameEntity(oldKeyName, newKeyName)
	if err != nil {
		fmt.Printf("Unable to rename key: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if !found {
		fmt.Printf("Unable to rename key: A key was not found with the name %s\n", oldKeyName)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = keystore.GlobalKeyStore.WriteToFile("")
	if err != nil {
		fmt.Printf("Unable to rename key: keystore could not update the file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Key renamed and keystore file changes committed.")
}
