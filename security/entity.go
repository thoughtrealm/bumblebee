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

package security

import "fmt"

type Entity struct {
	Name string
	Key  *KeyInfo
}

func (e *Entity) Print() {
	fmt.Printf("Name : %s\n", e.Key.Name)
	var value string
	switch e.Key.KeyType {
	case KeyTypePublic:
		value = string(e.Key.KeyData)
	case KeyTypeSeed:
		value = string(e.Key.KeyData)
	default:
		value = "UNKNOWN"
	}
	fmt.Printf("Type : %s\n", e.Key.KeyType.String())
	fmt.Printf("Value: %s\n", value)
	fmt.Println()
}

func (e *Entity) Clone() *Entity {
	return &Entity{
		Name: e.Name,
		Key:  e.Key.Clone(),
	}
}
