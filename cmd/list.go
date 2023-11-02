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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Shows lists of keys, keypairs, or profiles",
	Long:  "Shows lists of keys, keypairs, or profiles",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

type listSubCommandVals struct {
	match    string
	sort     bool
	pageSize int
}

var globalListSubCommandVals = &listSubCommandVals{}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringVarP(&globalListSubCommandVals.match, "match", "m", "", "A pattern for matching names. See docs for pattern.")
	listCmd.PersistentFlags().BoolVarP(&globalListSubCommandVals.sort, "sort", "s", false, "Indicates if results should be sorted or not.")
	listCmd.PersistentFlags().IntVarP(&globalListSubCommandVals.pageSize, "page-size", "p", 5, "Number of entities per page view.  If 0, not view break occurs.")
}
