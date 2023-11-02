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

// ExitCode is used as the final ExitCode return by this runtime
var ExitCode int = ExitCodeSuccess

// A list of ExitCode value mappings to specific errors.
// This is useful when running this utility from a shell call or some other app.
const (
	ExitCodeSuccess = iota
	ExitCodePanicInExecute
	ExitCodeErrorReturnedToExecute
	ExitCodeInvalidInput
	ExitCodeInputError
	ExitCodeCipherError
	ExitCodeStartupFailure
	ExitCodeOutputError
	ExitCodeRequestFailed
)
