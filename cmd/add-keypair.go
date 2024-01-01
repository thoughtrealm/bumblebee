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
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"

	"github.com/spf13/cobra"
)

// addKeypairCmd represents the keypair command
var addKeypairCmd = &cobra.Command{
	Use:   "keypair <name>",
	Args:  cobra.MaximumNArgs(1),
	Short: "Adds a new keypair identity",
	Long:  "Adds a new keypair identity",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		addNewKeyPair(args[0])
	},
}

func init() {
	addCmd.AddCommand(addKeypairCmd)

}

func addNewKeyPair(keypairName string) {
	var err error

	// This empty value for keypairName functionality was supported in the original implementation.
	// However, now, the keypairName should never be an empty string due to the logic in the calling
	// command.  Although, the user could use a construction like 'bumblebee add keypair ""', or something
	// else, that could result in an argument that is an empty string.
	// So, for now, instead of throwing an error, we will leave this support here, in case we want to
	// support this again in the future. Or, as a safeguard, if some other code should call this from
	// somewhere else without providing the keypairName value.
	if keypairName == "" {
		keypairName, err = helpers.GetConsoleRequiredInputLine(
			"Enter name for the new key",
			"Name",
		)

		if err != nil {
			fmt.Printf("Unable to get input for key name: %s\n", err)
			return
		}
	}

	kpi, err := keypairs.GlobalKeyPairStore.CreateNewKeyPair(keypairName)
	if err != nil {
		fmt.Printf("Unable to create new KeyPair: %s\n", err)
		return
	}

	err = kpi.Print("New Key Pair Info", true)
	if err != nil {
		fmt.Printf("unable to print keypair info: %s\n", errors.Unwrap(err))
		return
	}

	fmt.Println("KeyPair added")
	fmt.Println("Updating KeyPair Store file...")

	err = keypairs.GlobalKeyPairStore.SaveKeyPairStoreToOrigin(nil)
	if err != nil {
		fmt.Printf("unable to save keypair info: %s\n", errors.Unwrap(err))
		fmt.Println("New keypair was not saved to file")
		return
	}

	fmt.Println("New keypair stored to file")
}
