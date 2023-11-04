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

package keystore

type StoreDetails struct {
	Name    string
	Owner   string
	IsLocal bool
	IsDirty bool `msgpack:"-"`
}

func (sd *StoreDetails) Clone() *StoreDetails {
	return &StoreDetails{
		Name:    sd.Name,
		Owner:   sd.Owner,
		IsLocal: sd.IsLocal,
		IsDirty: sd.IsDirty,
	}
}

func (sks *SimpleKeyStore) GetDetails() *StoreDetails {
	if sks.Details == nil {
		return nil
	}

	return sks.Details.Clone()
}
