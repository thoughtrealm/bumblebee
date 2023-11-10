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
	"github.com/thoughtrealm/bumblebee/security"
)

// listKeypairsCmd represents the keypairs command
var listKeypairsCmd = &cobra.Command{
	Use:   "keypairs",
	Short: "Displays a list of all keypairs",
	Long:  "Displays a list of all keypairs",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		showKeyPairsList()
	},
}

type listKeyPairsSubCommandVals struct {
	showAll bool
}

var localListKeyPairsSubCommandVals = &listKeyPairsSubCommandVals{}

func init() {
	listCmd.AddCommand(listKeypairsCmd)
	listKeypairsCmd.Flags().BoolVarP(&localListKeyPairsSubCommandVals.showAll, "show-all", "", false, "Show all key elements for the keypair. Default is false and shows only the public key.")
}

// showKeysList will iterate the keys and display them
func showKeyPairsList() {
	if keypairs.GlobalKeyPairStore.Count() == 0 {
		fmt.Println("No key entities are loaded")
		return
	}

	fmt.Println("")
	fmt.Printf("Using profile   : %s\n", helpers.GlobalConfig.GetCurrentProfile().Name)
	fmt.Printf("KeyPairs Loaded : %d\n", keypairs.GlobalKeyPairStore.Count())
	fmt.Println("======================================================")
	keypairs.GlobalKeyPairStore.Walk(globalListSubCommandVals.sort, func(kpi *security.KeyPairInfo) {
		err := kpi.Print(
			"",
			localListKeyPairsSubCommandVals.showAll,
		)
		if err != nil {
			fmt.Printf("error printing keypair info: %s\n", err)
			return
		}

		fmt.Println()
	})
}
