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

type importCommandVals struct {
	inputSourceText string
	inputFilePath   string
	password        string
}

var sharedImportCommandVals = &importCommandVals{}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports users and keypairs from files, clipboard or piped input",
	Long:  "Imports users and keypairs from files, clipboard or piped input",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.PersistentFlags().StringVarP(&sharedImportCommandVals.inputSourceText, "input-source", "t", "file", "The input source.  Should be one of: pipe, clipboard or file.")
	importCmd.PersistentFlags().StringVarP(&sharedImportCommandVals.inputFilePath, "input-file", "f", "", "The file name to use for input. Only relevant if input-type is FILE.")
	importCmd.PersistentFlags().StringVarP(&sharedImportCommandVals.password,
		"password", "", "",
		`A password if required for the input stream.
If this is not provided and the input stream is password protected,
then you will be prompted for it.  Please be aware that providing 
passwords on the command line is not considered secure.
But if you are piping input or using bee in a pipe/process flow, then you can
use this flag to provide passwords for input streams as needed.`)
}
