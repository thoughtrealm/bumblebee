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

//go:build !(darwin && arm64) && !windows

package helpers

import (
	"errors"
)

func InitializeClipboard() (err error) {
	return errors.New("clipboard access is not available for this build of BumbleBee")
}

func WriteToClipboard(data []byte) error {
	return errors.New("clipboard access is not available for this build of BumbleBee")
}

func ReadFromClipboard() ([]byte, error) {
	return nil, errors.New("clipboard access is not available for this build of BumbleBee")
}
