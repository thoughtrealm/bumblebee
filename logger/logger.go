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

/*
	This is a really basic logger.  It suffices for the basic need of emitting debug output.
	Possibly, we will need some more sophisticated at some point.
*/

package logger

import (
	"fmt"
	"log"
	"os"
)

var Log = log.New(os.Stdout, "DEBUG: ", 0)

var Enabled bool
var LogTime bool = false
var logTimeChecked bool

func checkLogTime() {
	if logTimeChecked {
		return
	}

	logTimeChecked = true
	if LogTime {
		Log.SetFlags(log.Lmicroseconds)
	}
}

func Debug(text string) {
	if !Enabled {
		return
	}
	checkLogTime()
	Log.Println(text)
}

func Debugf(format string, a ...any) {
	if !Enabled {
		return
	}
	outputText := fmt.Sprintf(format, a...)
	Debug(outputText)
}
