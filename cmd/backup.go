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

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Will backup one or more profiles to a symmetrically encrypted file",
	Long:  "Will backup one or more profiles to a symmetrically encrypted file",
	Run: func(cmd *cobra.Command, args []string) {
		backupProfiles()
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

type backupCommandVals struct {
	// The user supplied key to decrypt the input with
	symmetricKey []byte

	// Command line provided symmetric key
	symmetricKeyInputText string

	// outputFile is the name of the backup file
	outputFile string

	// The name of the profile to backup.  Either a specific name or the word "all".
	profileName string
}

var localBackupCommandVals = &backupCommandVals{}

func backupProfiles() {

}
