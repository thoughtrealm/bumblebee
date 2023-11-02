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
	"github.com/thoughtrealm/bumblebee/env"
)

// listProfilesCmd represents the profile command
var addProfileCmd = &cobra.Command{
	Use:   "profile",
	Args:  cobra.MaximumNArgs(1),
	Short: "Adds a new profile",
	Long:  "Adds a new profile",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		var profileName string
		if len(args) == 1 {
			profileName = args[0]
		}

		addNewProfile(profileName)
	},
}

func init() {
	addCmd.AddCommand(addProfileCmd)
}

func addNewProfile(profileName string) {
	err := env.CreateNewProfile(profileName)
	if err != nil {
		fmt.Printf("Failed creating new profile: %s\n", err)
	}
}
