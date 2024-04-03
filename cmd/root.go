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
	"github.com/thoughtrealm/bumblebee/bootstrap"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/keypairs"
	logger "github.com/thoughtrealm/bumblebee/logger"
	"os"
)

type rootCommandVals struct {
	useProfile      string
	outputValueOnly bool
}

var sharedRootCommandVals = &rootCommandVals{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bumblebee",
	Short: "Bumblebee - A utility for sharing secrets",
	Long:  "Bumblebee - A utility for sharing secrets",
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args) == 1 && !helpers.CheckIsPiped() {
			_ = cmd.Help()
			return
		}

		if helpers.CheckIsPiped() {
			// what to do here? are we going to support piping?
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		if r := recover(); r != nil {
			helpers.ExitCode = helpers.ExitCodePanicInExecute
			// do want to honor the no output concept for piping as well relating to critical errors?
			if !helpers.CmdHelpers.OutputValueOnly {
				fmt.Printf("Panic recovered in cmd.Execute(): %s\n", r)
			}
		}
	}()

	defer keypairs.WipeGlobalKeyPairsIfValid()

	err := rootCmd.Execute()
	if err != nil {
		helpers.ExitCode = helpers.ExitCodeErrorReturnedToExecute
	}
}

func init() {
	cmd := GetRootCmd()
	/* copied from another project for example in future changes
	cmd.Example = "  u2t 681678000\n" +
		"  u2t 681678000 --output-format RFC3339\n" +
		"  u2t 681678000000 --input-type milli --input-format secs --output-format RFC3339\n" +
		"  u2t 681678000000 --input-type milli --output-format custom --custom-text \"mmm yyyy-mm-dd hhh:nn:ss.000 zthhmm\n" +
		"  u2t 681678000000 --input-type milLI --output-format customGo --custom-text \"Jan 2006-01-02 15:04:05.000 Z-0700\n" +
		"  u2t list --output-formats\n" +
		"  u2t list --custom-entities"
	*/

	cmd.PersistentFlags().BoolVarP(&helpers.CmdHelpers.OutputValueOnly, "output-only", "v", false, "Only print necessary output.  This is usually for removing extraneous characters\nwhen piping output to another process.")
	cmd.PersistentFlags().BoolVarP(&logger.LogDebug, "log-debug", "g", false, "Enables debug logging output")
	cmd.PersistentFlags().BoolVarP(&logger.LogTime, "log-time", "", false, "Adds time to debug lines.  Only relevant if debug output is enabled with \"--log\"")
	cmd.PersistentFlags().BoolVarP(&logger.LogDebugVerbose, "log-debug-verbose", "", false, "Enables output of detailed debug information.")
	cmd.PersistentFlags().StringVarP(&helpers.CmdHelpers.UseProfile, "use", "u", "", "The name of the profile to use for the specified command.")
}

func ShowUsage(cmd *cobra.Command) error {
	fmt.Println("I am a usage func.  Verify me.")
	fmt.Println(cmd.Use)
	return nil
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func startBootStrap(loadKeystore, loadKeypairStore bool) error {
	err := bootstrap.Run(loadKeystore, loadKeypairStore)
	if err != nil {
		fmt.Printf("Failure loading bumblebee%s\n", helpers.FormatErrorOutputs(err))
		return err
	}

	return nil
}
