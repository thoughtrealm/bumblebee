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

type InputType int

const (
	InputTypeConsole InputType = iota
	InputTypeFile
	InputTypeClipboard
	InputTypePiped
	InputTypeUnknown
)

func TextToInputType(textName string) InputType {
	switch strings.ToUpper(strings.Trim(textName, " \t\n\r")) {
	case "CONSOLE":
		return InputTypeConsole
	case "FILE":
		return InputTypeFile
	case "CLIPBOARD":
		return InputTypeClipboard
	case "PIPED":
		return InputTypePiped
	default:
		return InputTypeUnknown
	}
}

type OutputType int

const (
	OutputTypeConsole OutputType = iota
	OutputTypeFile
	OutputTypePath
	OutputTypeClipboard
	OutputTypeUnknown
)

func TextToOutputType(textName string) OutputType {
	switch strings.ToUpper(strings.Trim(textName, " \t\n\r")) {
	case "CONSOLE":
		return OutputTypeConsole
	case "FILE":
		return OutputTypeFile
	case "PATH":
		return OutputTypePath
	case "CLIPBOARD":
		return OutputTypeClipboard
	default:
		return OutputTypeUnknown
	}
}
