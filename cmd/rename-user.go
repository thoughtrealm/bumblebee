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

// renameUserCmd represents the key command
var renameUserCmd = &cobra.Command{
	Use:   "user <oldName> <newName>",
	Args:  cobra.ExactArgs(2),
	Short: "Will change the name of a user from oldName to newName",
	Long:  "Will change the name of a user from oldName to newName",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Exact args is set by Args property above. So we should ALWAYS have 2 args
		oldUserName := args[0]
		newUserName := args[1]

		renameUser(oldUserName, newUserName)
	},
}

func init() {
	renameCmd.AddCommand(renameUserCmd)
}

func renameUser(oldUserName, newUserName string) {
	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to rename user: keystore not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	found, err := keystore.GlobalKeyStore.RenameEntity(oldUserName, newUserName)
	if err != nil {
		fmt.Printf("Unable to rename user: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if !found {
		fmt.Printf("Unable to rename user: A user was not found with the name %s\n", oldUserName)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	err = keystore.GlobalKeyStore.WriteToFile("")
	if err != nil {
		fmt.Printf("Unable to rename user: keystore could not update the file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("User renamed and keystore file changes committed.")
}
