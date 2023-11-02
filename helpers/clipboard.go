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

package helpers

import (
	"fmt"
	"golang.design/x/clipboard"
)

var clipboardInitialized bool

func InitializeClipboard() error {
	if clipboardInitialized {
		return nil
	}

	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("unable to initialize clipboard: %w", err)
	}

	clipboardInitialized = true
	return nil
}

func WriteToClipboard(data []byte) error {
	err := InitializeClipboard()
	if err != nil {
		return err
	}

	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func ReadFromClipboard() ([]byte, error) {
	err := InitializeClipboard()
	if err != nil {
		return nil, err
	}

	data := clipboard.Read(clipboard.FmtText)
	return data, nil
}
