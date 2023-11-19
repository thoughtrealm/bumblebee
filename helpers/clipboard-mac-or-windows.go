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

//go:build (darwin && arm64) || windows

package helpers

import (
	"errors"
	"fmt"
	"golang.design/x/clipboard"
	"strings"
)

var clipboardInitialized bool

func InitializeClipboard() (err error) {
	if clipboardInitialized {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			errText := fmt.Sprintf("%s", r)
			if strings.Contains(strings.ToLower(errText), "clipboard: cannot use when cgo_enabled=0") {
				err = errors.New("clipboard access is not available for this build of BumbleBee")
				return
			}

			err = fmt.Errorf("%s", r)
		}
	}()

	err = clipboard.Init()
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
