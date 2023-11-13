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
	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user [--password] [--input-source] [--input-file]",
	Short: "Imports user info into the keystore from files, clipboard or piped input",
	Long:  "Imports user info into the keystore from files, clipboard or piped input",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// bootstrap will print its own messages
			return
		}

		importUser()
	},
}

func init() {
	importCmd.AddCommand(userCmd)
}

func importUser() {

}
