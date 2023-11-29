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
)

// Some of these global vars are placeholders for the build command, when
// relevant values are inserted.

var (
	AppName         = "Bumblebee - A utility for sharing secrets"
	AppMajorVersion = "0"
	AppMinorVersion = "1"
	AppPatchVersion = "0"

	// For any pre-release version, it would need to provide leading ".", like ".dev01"
	AppPreReleaseVer  = "-alpha"
	AppVersion        = AppMajorVersion + "." + AppMinorVersion + "." + AppPatchVersion + AppPreReleaseVer
	AppShortBuildTime = "[sbt]"
	AppLongBuildTime  = "[lbt]"
	AppProject        = "GITHUB https://github.com/thoughtrealm/bumblebee"
	AppLicense        = "MIT License https://github.com/thoughtrealm/bumblebee/blob/main/LICENSE"
)

func printVersionInfo(inPromptMode bool) {
	fmt.Println("")
	fmt.Printf("%s\n\n", AppName)
	fmt.Printf("Version          : %s\n", AppVersion)
	fmt.Printf("Build Time[short]: %s\n", AppShortBuildTime)
	fmt.Printf("Build Time[long] : %s\n", AppLongBuildTime)
	fmt.Printf("Project          : %s\n", AppProject)
	fmt.Printf("License          : %s\n", AppLicense)
	fmt.Println("")
}
