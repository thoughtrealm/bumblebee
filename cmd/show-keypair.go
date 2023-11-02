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

// showKeyPairCmd represents the keyPair command
var showKeyPairCmd = &cobra.Command{
	Use:   "keypair [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Will display the keypair data for the referenced name",
	Long:  "Will display the keypair data for the referenced name",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		var keyPairName string
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		keyPairName = args[0]

		showKeypair(keyPairName)
	},
}

var showKeypairInfoShowAll bool

func init() {
	showCmd.AddCommand(showKeyPairCmd)
	showKeyPairCmd.Flags().BoolVarP(&showKeypairInfoShowAll, "show-all", "a", false, "Will output all key info, not just the public key")
}

func showKeypair(keypairName string) {
	if keypairs.GlobalKeyPairStore == nil {
		fmt.Println("Unable to show keypair.  Keypair store not loaded")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	kpi := keypairs.GlobalKeyPairStore.GetKeyPairInfo(keypairName)
	if kpi == nil {
		fmt.Printf("No keypair with the name \"%s\" was located\n", keypairName)
		return
	}

	fmt.Printf("Using profile: %s\n", helpers.GlobalConfig.GetCurrentProfile().Name)
	err := kpi.Print("Keypair Info", showKeypairInfoShowAll)
	if err != nil {
		fmt.Printf("Unable to print keypair info: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
}
