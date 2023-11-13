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
	"github.com/thoughtrealm/bumblebee/helpers"
)

type exportCommandVals struct {
	exportOutputTargetText   string
	exportOutputTarget       helpers.ExportOutputTarget
	exportOutputFilePath     string
	exportOutputEncodingText string
	exportOutputEncoding     helpers.ExportOutputEncoding
	exportPassword           string
}

var sharedExportCommandVals = &exportCommandVals{}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Will export user or keypair info for adding to another profile or system",
	Long:  "Will export user or keypair info for adding to another profile or system",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().StringVarP(&sharedExportCommandVals.exportOutputTargetText, "output-target", "t", "console", "The output target.  Should be one of: console, clipboard or file.")
	exportCmd.PersistentFlags().StringVarP(&sharedExportCommandVals.exportOutputFilePath, "output-file", "f", "", "The file name to use for output. Only relevant if output-type is FILE.")
	exportCmd.PersistentFlags().StringVarP(&sharedExportCommandVals.exportOutputEncodingText, "output-encoding", "e", "text", "The encoding for the output.  Should be \"text\" or \"raw\".\nText is human readable and can be copied, printed, pasted into an email, texted, etc.\nRaw is not readable or printable to console, text docs, emails, etc.")
	exportCmd.PersistentFlags().StringVarP(&sharedExportCommandVals.exportPassword,
		"password", "", "",
		`A password to use for encrypting the output stream.
If this is not provided, you will be prompted for the password.
This flag should only be used when piping export info to another process,
since passing passwords on the command line is not considered secure.
For user exports, the password is optional, but for keypair exports, 
the password is required.
So if you wish to export keypair info to another process, 
you must use this flag to provide the password.`)
}

func sharedProcessExportFlags() {
	sharedExportCommandVals.exportOutputTarget = helpers.TextToExportOutputTarget(
		sharedExportCommandVals.exportOutputTargetText)

	sharedExportCommandVals.exportOutputEncoding = helpers.TextToExportOutputEncoding(
		sharedExportCommandVals.exportOutputEncodingText)
}
