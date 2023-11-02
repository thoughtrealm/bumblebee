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
	"os"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Sets the referenced profile as the default profile",
	Long:  "Sets the referenced profile as the default profile",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		useProfile()
	},
}

type useCommandVals struct {
	name string
}

var localUseCommandVals = &useCommandVals{}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().StringVarP(&localUseCommandVals.name, "name", "n", "", "The name of the profile to select")
}

func useProfile() {
	if len(os.Args) < 2 {
		fmt.Println("Insufficient arguments: expected bee use [profile_name]")
		return
	}

	name := os.Args[2]

	if helpers.GlobalConfig == nil {
		fmt.Println("Config not loaded")
		return
	}

	profile := helpers.GlobalConfig.GetProfile(name)
	if profile == nil {
		fmt.Printf("No profile was located for name \"%s\"\n", name)
		return
	}

	helpers.GlobalConfig.Config.CurrentProfile = name

	err := helpers.GlobalConfig.WriteConfig()
	if err != nil {
		fmt.Printf("Unable to save new config: %s\n", err)
		return
	}

	fmt.Printf("Default profile updated to \"%s\"\n", name)
}
