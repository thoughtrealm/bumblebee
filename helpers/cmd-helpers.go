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

// HelpersInfo is used to store the command line parameters, as well as transformed values.
// These are set by the cmd package/cobra, and are used or transformed in the primary logic in the encode package.
type HelpersInfo struct {
	OutputValueOnly bool
	PipeMode        bool
	UseProfile      string
}

var CmdHelpers = &HelpersInfo{}
