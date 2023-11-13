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
	"time"
)

var Log = log.New(os.Stdout, "DEBUG: ", 0)

var (
	LogDebug         = false
	LogDebugVerbose  = false
	LogTime          = false
	LogOutputOnly    = false
	logConfigChecked = false
	stdoutTarget     = os.Stdout
	stderrTarget     = os.Stderr
	debugoutTarget   = os.Stdout
)

type logMode int

const (
	logModeNormal logMode = iota
	logModeOutputOnly
	logModeDebug
	logModeDebugVerbose
)

// mode ... By default, assume we want to use a normal logmode, which is no debugging and allow user info.
var mode logMode = logModeNormal

func checkLogConfig() {
	if logConfigChecked {
		return
	}

	logConfigChecked = true

	if LogOutputOnly {
		mode = logModeOutputOnly
		return
	}

	if LogDebugVerbose {
		mode = logModeDebugVerbose
		return
	}

	if LogDebug {
		mode = logModeDebug
		return
	}

	// no explicit log mode requests are set, so we just use normal, which is no debugging and allow user info
	mode = logModeNormal
}

func buildDebugPrefix() string {
	if !LogTime {
		return "DEBUG: "
	}

	timeFormatted := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("DEBUG  %s : ", timeFormatted)
}

func outputLn(target *os.File, prefix, text string) {
	_, _ = fmt.Fprintln(target, prefix+text)
}

func outputfLn(target *os.File, prefix, format string, a ...any) {
	_, _ = fmt.Fprint(target, prefix+fmt.Sprintf(format, a...))
}

func Debug(text string) {
	checkLogConfig()
	if mode != logModeDebug && mode != logModeDebugVerbose {
		return
	}

	outputLn(debugoutTarget, buildDebugPrefix(), text)
}

func Debugf(format string, a ...any) {
	checkLogConfig()
	if mode != logModeDebug && mode != logModeDebugVerbose {
		return
	}

	outputfLn(debugoutTarget, buildDebugPrefix(), format, a...)
}

func DebugVerbose(text string) {
	checkLogConfig()
	if mode != logModeDebugVerbose {
		return
	}

	outputLn(debugoutTarget, buildDebugPrefix(), text)
}

func DebugVerbosef(format string, a ...any) {
	checkLogConfig()
	if mode != logModeDebugVerbose {
		return
	}

	outputfLn(debugoutTarget, buildDebugPrefix(), format, a...)
}

func Println(text string) {
	checkLogConfig()
	if mode == logModeOutputOnly {
		return
	}

	outputLn(stdoutTarget, "", text)
}

func Printf(format string, a ...any) {
	checkLogConfig()
	if mode == logModeOutputOnly {
		return
	}

	outputfLn(stdoutTarget, "", format, a...)
}

// Output ALWAYS PRINTS
func Output(text string) {
	outputLn(stdoutTarget, "", text)
}

// Outputf ALWAYS PRINTS
func Outputf(format string, a ...any) {
	outputfLn(stdoutTarget, "", format, a...)
}

// Error ALWAYS PRINTS
func Error(text string) {
	outputLn(stderrTarget, "", text)
}

// Errorf ALWAYS PRINTS
func Errorf(format string, a ...any) {
	outputfLn(stderrTarget, "", format, a...)
}
