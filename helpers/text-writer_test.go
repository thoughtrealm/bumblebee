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

import "testing"

func TestBytesScreenWriterHex_Simple10x10(t *testing.T) {
	testBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bsw := NewTextWriter(TextWriterTargetConsole, 10, TextWriterModeBinary, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}

func TestBytesScreenWriterHex_Width20_6x10(t *testing.T) {
	testBytes := []byte{1, 2, 3, 4, 5, 6}
	bsw := NewTextWriter(TextWriterTargetConsole, 20, TextWriterModeBinary, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}

func TestBytesScreenWriterHex_Width20_7x10(t *testing.T) {
	testBytes := []byte{1, 2, 3, 4, 5, 6, 7}
	bsw := NewTextWriter(TextWriterTargetConsole, 20, TextWriterModeBinary, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}

func TestBytesScreenWriterText_Simple10x10(t *testing.T) {
	testBytes := []byte("1234567890")
	bsw := NewTextWriter(TextWriterTargetConsole, 10, TextWriterModeText, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}

func TestBytesScreenWriterText_Width20_6x10(t *testing.T) {
	testBytes := []byte("123456")
	bsw := NewTextWriter(TextWriterTargetConsole, 20, TextWriterModeText, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}

func TestBytesScreenWriterText_Width20_7x10(t *testing.T) {
	testBytes := []byte("1234567")
	bsw := NewTextWriter(TextWriterTargetConsole, 20, TextWriterModeText, "", "", nil, nil)
	for x := 0; x < 10; x++ {
		_, _ = bsw.Write(testBytes)
	}
	bsw.Flush()
}
