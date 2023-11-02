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

// showConfigCmd represents the config command
var showConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Displays the config definitions",
	Long:  "Displays the config definitions",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(false, false)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		showConfig()
	},
}

func init() {
	showCmd.AddCommand(showConfigCmd)
}

func showConfig() {
	if helpers.GlobalConfig == nil {
		fmt.Println("Unable to show config: global config not loaded")
		return
	}

	yamlLines, err := helpers.GlobalConfig.GetConfigYAMLLines()
	if err != nil {
		fmt.Printf("Unable to get config data: %s\n", errors.Unwrap(err))
		return
	}

	fmt.Println("Full Config Definition")
	fmt.Println("===========================================================================")
	for _, line := range yamlLines {
		fmt.Println(line)
	}
}
