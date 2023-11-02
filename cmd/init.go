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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a Bumblebee environment",
	Long:  "Initializes a Bumblebee environment",
	Run: func(cmd *cobra.Command, args []string) {
		err := env.InitializeEnvironment()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		fmt.Println("New default environment created successfully")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
