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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceFileExt(t *testing.T) {
	type TestVals struct {
		Name      string
		InputPath string
		NewExt    string
		Expects   string
	}

	tests := []TestVals{
		{
			Name:      "FileNameAndNewExtHasPeriod",
			InputPath: "file.old",
			NewExt:    ".new",
			Expects:   "file.new",
		},
		{
			Name:      "FileNameAndNewExtHasNoPeriod",
			InputPath: "file.old",
			NewExt:    "new",
			Expects:   "file.new",
		},
		{
			Name:      "FileNameWithNoExtAndNewExtHasPeriod",
			InputPath: "file",
			NewExt:    ".new",
			Expects:   "file.new",
		},
		{
			Name:      "FileNameWithNoExtAndNewExtHasNoPeriod",
			InputPath: "file",
			NewExt:    "new",
			Expects:   "file.new",
		},
		{
			Name:      "FilePathWithMultiplePeriodsAndNewExtHasPeriod",
			InputPath: "/path/subpath/file.something.old",
			NewExt:    ".new",
			Expects:   "/path/subpath/file.something.new",
		},
		{
			Name:      "FilePathWithMultiplePeriodsAndNewExtHasNoPeriod",
			InputPath: "/path/subpath/file.something.old",
			NewExt:    "new",
			Expects:   "/path/subpath/file.something.new",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			value := ReplaceFileExt(test.InputPath, test.NewExt)
			assert.Equal(t, test.Expects, value)
		})
	}
}
