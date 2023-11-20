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

package keystore

import "strings"

type BundleType int

const (
	BundleTypeCombined BundleType = iota
	BundleTypeSplit
	BundleTypeUnknown
)

func TextToBundleType(textName string) BundleType {
	switch strings.ToUpper(strings.Trim(textName, " \t\n\r")) {
	case "COMBINED":
		return BundleTypeCombined
	case "SPLIT":
		return BundleTypeSplit
	default:
		return BundleTypeUnknown
	}
}

type InputSource int

const (
	InputSourceConsole InputSource = iota
	InputSourceFile
	InputSourceClipboard
	InputSourcePiped
	InputSourceUnknown
)

func TextToInputSource(textName string) InputSource {
	switch strings.ToUpper(strings.Trim(textName, " \t\n\r")) {
	case "CONSOLE":
		return InputSourceConsole
	case "FILE":
		return InputSourceFile
	case "CLIPBOARD":
		return InputSourceClipboard
	case "PIPED":
		return InputSourcePiped
	default:
		return InputSourceUnknown
	}
}

type OutputTarget int

const (
	OutputTargetConsole OutputTarget = iota
	OutputTargetFile
	OutputTargetPath
	OutputTargetClipboard
	OutputTargetUnknown
)

func TextToOutputTarget(textName string) OutputTarget {
	switch strings.ToUpper(strings.Trim(textName, " \t\n\r")) {
	case "CONSOLE":
		return OutputTargetConsole
	case "FILE":
		return OutputTargetFile
	case "PATH":
		return OutputTargetPath
	case "CLIPBOARD":
		return OutputTargetClipboard
	default:
		return OutputTargetUnknown
	}
}
