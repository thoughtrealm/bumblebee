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
	"github.com/spf13/cobra"
	"github.com/thoughtrealm/bumblebee/helpers"
)

// showProfileCmd represents the profile command
var showProfileCmd = &cobra.Command{
	Use:   "profile [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Displays profile metadata",
	Long:  "Displays profile metadata. If name is omitted, it shows the current default profile.",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		}

		showProfile(profileName)
	},
}

func init() {
	showCmd.AddCommand(showProfileCmd)
}

func showProfile(profileName string) {
	if helpers.GlobalConfig == nil {
		fmt.Println("Unable to show profile: global config not loaded")
		return
	}

	yamlLines, profileNameOut, err := helpers.GlobalConfig.GetProfileYAMLLines(profileName)
	if err != nil {
		fmt.Printf("Unable to get profile config data: %s\n", errors.Unwrap(err))
		return
	}

	fmt.Printf("Config Definition for Profile: %s\n", profileNameOut)
	fmt.Println("===========================================================================")
	for _, line := range yamlLines {
		fmt.Println(line)
	}
}
