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
	"strings"
)

// removeProfileCmd represents the profile command
var removeProfileCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Removes the referenced profile and all related data",
	Long:  "Removes the referenced profile and all related data",
	Run: func(cmd *cobra.Command, args []string) {
		// Only load the config YAML data
		err := startBootStrap(false, false)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		var profileName string
		switch len(args) {
		case 0:
			_ = cmd.Help()
			return
		case 1:
			profileName = args[0]
		}

		removeProfile(profileName)
	},
}

func init() {
	removeCmd.AddCommand(removeProfileCmd)
}

func removeProfile(profileName string) {
	if strings.ToUpper(profileName) == "DEFAULT" {
		fmt.Println("You may not remove the DEFAULT profile.  Instead, you can re-init the environment with \"bee init\".")
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	if helpers.GlobalConfig == nil {
		fmt.Println("Unable to remove profile: config not available")
		helpers.ExitCode = helpers.ExitCodeStartupFailure
		return
	}

	profile := helpers.GlobalConfig.GetProfile(profileName)
	if profile == nil {
		fmt.Printf("A profile was not found with name \"%s\"\n", profileName)
		return
	}

	fmt.Println("Removing this profile will delete all associated keys and keypairs.")
	fmt.Println("")
	response, err := helpers.GetYesNoInput(
		fmt.Sprintf("Are you sure you wish to remove the profile \"%s\"?", profileName),
		helpers.InputResponseValNo,
	)
	if err != nil {
		fmt.Printf("Unable to confirm removal of profile \"%s\"\n due to an error: %s", profileName, err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	if response != helpers.InputResponseValYes {
		fmt.Println("User aborted removal request")
		helpers.ExitCode = helpers.ExitCodeInputError
		return
	}

	// First, remove the profile's path under global config
	err = os.RemoveAll(profile.Path)
	if err != nil {
		fmt.Printf("Unable to remove profile path: %s\n", err)
		fmt.Printf("Profile path: \"%s\"\n", profile.Path)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	// Now, remove the profile ref from config metadata
	found, err := helpers.GlobalConfig.RemoveProfile(profileName)
	if err != nil {
		fmt.Printf("Unable to remove profile reference from config metadata: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}
	if !found {
		fmt.Println("Profile reference was not found in config metadata")
		return
	}

	err = helpers.GlobalConfig.WriteConfig()
	if err != nil {
		fmt.Printf("Unable to update the config file: %s\n", err)
		helpers.ExitCode = helpers.ExitCodeRequestFailed
		return
	}

	fmt.Println("Profile artifacts removed and config file updated")
}
