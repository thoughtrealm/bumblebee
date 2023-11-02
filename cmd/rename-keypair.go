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
	"github.com/thoughtrealm/bumblebee/keypairs"
	"strings"
)

// renameKeypairCmd represents the keypair command
var renameKeypairCmd = &cobra.Command{
	Use:   "keypair <oldName> <newName>",
	Short: "Will change the name of a keypair from oldName to newName",
	Long:  "Will change the name of a keypair from oldName to newName",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		// Exact args is set by Args property above. So we should ALWAYS have 2 args
		oldKeypairName := args[0]
		newKeypairName := args[1]

		renameKeypair(oldKeypairName, newKeypairName)
	},
}

func init() {
	renameCmd.AddCommand(renameKeypairCmd)
}

func renameKeypair(oldKeypairName, newKeypairName string) {
	if strings.ToLower(oldKeypairName) == "default" {
		fmt.Println("Renaming keypair \"default\" is not allowed.")
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	if keypairs.GlobalKeyPairStore == nil {
		fmt.Println("Unable to rename keypair: keypair store not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	found, err := keypairs.GlobalKeyPairStore.RenameKeyPair(oldKeypairName, newKeypairName)
	if err != nil {
		fmt.Printf("Unable to rename keypair: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if !found {
		fmt.Printf("Unable to rename keypair: keypair not found with name \"%s\"\n", oldKeypairName)
		helpers.ExitCode = helpers.ExitCodeInvalidInput
		return
	}

	err = keypairs.GlobalKeyPairStore.SaveKeyPairStoreToOrigin(nil)
	if err != nil {
		fmt.Printf("Unable to rename keypair: keypair store could not update the file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Keypair renamed and keypair store file changes committed.")
}
