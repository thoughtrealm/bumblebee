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
	"github.com/thoughtrealm/bumblebee/security"
)

// listKeysCmd represents the keys command
var listKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Displays a list of all keys",
	Long:  "Displays a list of all keys",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		showKeysList()
	},
}

func init() {
	listCmd.AddCommand(listKeysCmd)
}

// showKeysList will iterate the keys and display them
func showKeysList() {
	if keystore.GlobalKeyStore.Count() == 0 {
		fmt.Println("No key entities are loaded")
		return
	}

	// First, walk the map to get a count of matching items
	walkCount, err := keystore.GlobalKeyStore.WalkCount(globalListSubCommandVals.match, nil)
	if err != nil {
		fmt.Printf("Failed on pass 1 entity count: %s\n", err)
	}

	fmt.Println("")
	fmt.Printf("Using profile : %s\n", helpers.GlobalConfig.GetCurrentProfile().Name)
	fmt.Printf("Keys Loaded   : %d, Keys Matched: %d\n", keystore.GlobalKeyStore.Count(), walkCount)
	fmt.Println("======================================================")

	// instead of passing in an inline func declaration, we'll use a var so the call is more readable
	walkFunc := func(entity *security.Entity) {
		entity.Print()
	}

	walkInfo := keystore.NewWalkInfo(globalListSubCommandVals.match, globalListSubCommandVals.sort, nil, walkFunc)

	_ = keystore.GlobalKeyStore.Walk(walkInfo)
}
