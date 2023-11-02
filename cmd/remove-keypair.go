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
)

// removeKeypairCmd represents the keypair command
var removeKeypairCmd = &cobra.Command{
	Use:   "keypair [name]",
	Short: "Will remove the referenced keypair from the keypair store",
	Long:  "Will remove the referenced keypair from the keypair store",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		var keypairName string
		switch len(args) {
		case 0:
			_ = cmd.Help()
			return
		case 1:
			keypairName = args[0]
		}

		removeKeypair(keypairName)
	},
}

var removeKeypairInfoShowAll bool

func init() {
	removeCmd.AddCommand(removeKeypairCmd)
	removeKeypairCmd.Flags().BoolVarP(&removeKeypairInfoShowAll, "show-all", "a", false, "Will output all key info, not just the public key")
}

func removeKeypair(keypairName string) {
	if keypairs.GlobalKeyPairStore == nil {
		fmt.Println("Unable to remove keypair: keypair store not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	kpi := keypairs.GlobalKeyPairStore.GetKeyPairInfo(keypairName)
	if kpi == nil {
		fmt.Printf("No keypair was found with name \"%s\"\n", keypairName)
		return
	}

	fmt.Println("Keypair info located...")
	err := kpi.Print("Keypair Info", removeKeypairInfoShowAll)
	if err != nil {
		fmt.Printf("Unable to display requested keypair info: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
	fmt.Println()

	response, err := helpers.GetYesNoInput(fmt.Sprintf("Are you sure you wish to remove the keypair \"%s\"?", keypairName), helpers.InputResponseValNo)
	if err != nil {
		fmt.Printf("Unable to confirm removal of keypair: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if response != helpers.InputResponseValYes {
		fmt.Println("User aborted removal request")
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	found, err := keypairs.GlobalKeyPairStore.RemoveKeyPair(keypairName)
	if err != nil {
		fmt.Printf("Unable to remove keypair: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if !found {
		fmt.Printf("The keypair was not found during removal: %s\n", keypairName)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Keypair removed and keypair store updated")
}
