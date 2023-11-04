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

package cmd

import (
	"testing"
)

func TestListProfiles(t *testing.T) {
	//outputBuff := bytes.NewBuffer(nil)
	//rootCmd.SetOut(outputBuff)
	rootCmd.SetArgs([]string{"show", "profile", "alice"})
	_ = rootCmd.Execute()
	//fmt.Println("captured output")
	//fmt.Println("===========================")
	//fmt.Fprint(os.Stdout, outputBuff)
}
