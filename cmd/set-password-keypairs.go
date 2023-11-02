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
	"github.com/thoughtrealm/bumblebee/keypairs"
)

// passwordKeypairsCmd represents the keypairs subcommand for "set password" command
var passwordKeypairsCmd = &cobra.Command{
	Use:   "keypairs",
	Short: "Sets the password for the keypair store",
	Long:  "Sets the password for the keypair store",
	Run: func(cmd *cobra.Command, args []string) {
		err := startBootStrap(true, true)
		if err != nil {
			// startBootstrap prints messages, so nothing to print here, just bail
			return
		}

		setKeyPairsPassword()
	},
}

func init() {
	passwordCmd.AddCommand(passwordKeypairsCmd)
}

func setKeyPairsPassword() {
	fmt.Println("Enter a new password for the keypair store. Enter an empty value if you wish to not have a password.")
	fmt.Println("")
	fmt.Printf("New password: ")

	newPasswordBytes, err := helpers.GetPasswordWithConfirm("")
	if err != nil {
		fmt.Printf("Unable to obtain new password: %s\n", err)
		return
	}

	if len(newPasswordBytes) == 0 {
		fmt.Println("You have entered an empty password value.  Your keypair store will be unencrypted.")
		response, err := helpers.GetYesNoInput("Are you sure you wish to leave your keypair store unencrypted", helpers.InputResponseValNo)
		if err != nil {
			fmt.Printf("Unable to validate empty password: %s\n", err)
			return
		}

		if response != helpers.InputResponseValYes {
			fmt.Println("set password keypairs\" aborted.")
			return
		}
	}

	keypairs.GlobalKeyPairStore.SetPassword(newPasswordBytes)
	fmt.Println("In-memory keypair store password updated.")

	fmt.Println("Saving keypair store with updated password...")
	err = keypairs.GlobalKeyPairStore.SaveKeyPairStoreToOrigin(newPasswordBytes)
	if err != nil {
		fmt.Printf("Unable to update keypair store file: %s\n", err)
		return
	}

	fmt.Println("Keypair store file updated with new password.")

	fmt.Println("Updating system config metadata...")
	profile := helpers.GlobalConfig.GetCurrentProfile()
	if profile == nil {
		fmt.Println("Unable to retrieve current profile config")
		fmt.Println("Config keystore encryption state may be out of sync with keystore file")
		return
	}

	profile.KeyPairStoreEncrypted = (len(newPasswordBytes) > 0)
	err = helpers.GlobalConfig.WriteConfig()
	if err != nil {
		fmt.Printf("Unable to write updating config metadata: %s\n", err)
		return
	}

	fmt.Println("Config metadata updated")
}
