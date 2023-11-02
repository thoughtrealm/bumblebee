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
	"time"
)

func FormatDuration(totalTime time.Duration) string {
	if totalTime == 0 {
		return "0 seconds"
	}

	if totalTime/time.Millisecond == 0 {
		return "0 milliseconds"
	}

	if totalTime/time.Second == 0 {
		return fmt.Sprintf("%d milliseconds", totalTime/time.Millisecond)
	}

	seconds := totalTime / time.Second
	milliseconds := (totalTime - (seconds * time.Second)) / time.Millisecond
	return fmt.Sprintf("%d.%03d secs", seconds, milliseconds)
}
