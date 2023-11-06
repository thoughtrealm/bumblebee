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
	"github.com/thoughtrealm/bumblebee/helpers"
	"sort"

	"github.com/spf13/cobra"
)

// listProfilesCmd represents the profile command
var listProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Displays a list of all profiles",
	Long:  "Displays a list of all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		listProfiles()
	},
}

func init() {
	listCmd.AddCommand(listProfilesCmd)
}

func listProfiles() {
	if helpers.GlobalConfig == nil {
		fmt.Println("Unable to list profiles: global config not loaded")
		return
	}

	if len(helpers.GlobalConfig.Config.Profiles) == 0 {
		fmt.Println("No profiles currently exist in the global config metadata")
		return
	}

	configClone := helpers.GlobalConfig.Config.Clone()
	sort.Slice(configClone.Profiles, func(i, j int) bool {
		if helpers.CompareStrings(configClone.Profiles[i].Name, configClone.Profiles[j].Name) == -1 {
			return true
		}

		return false
	})

	fmt.Println("")
	fmt.Printf("Total Profiles Loaded: %d\n", len(configClone.Profiles))
	fmt.Println("============================================================")
	for idx, profile := range configClone.Profiles {
		fmt.Printf("Profile %2d: %s\n", idx+1, profile.Name)
	}
	fmt.Println("")
}
