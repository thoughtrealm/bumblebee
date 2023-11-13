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

// showUserCmd represents the user subcommand
var showUserCmd = &cobra.Command{
	Use:   "user [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Will display the referenced user info",
	Long:  "Will display the referenced user info",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		var keyName string
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		keyName = args[0]

		showUser(keyName)
	},
}

func init() {
	showCmd.AddCommand(showUserCmd)
}

func showUser(userName string) {
	if keystore.GlobalKeyStore == nil {
		fmt.Println("Unable to show user info: Key Store not loaded.")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	entity := keystore.GlobalKeyStore.GetKey(userName)
	if entity == nil {
		fmt.Printf("No user with the name \"%s\" was found.\n", userName)
		return
	}

	fmt.Printf("Using profile: %s\n", helpers.GlobalConfig.GetCurrentProfile().Name)
	entity.Print()
}
