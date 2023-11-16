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

import "github.com/thoughtrealm/bumblebee/security"

// This initial, basic server-related functionality is for the future implementation of a key management service.
// It is not currently implemented or used, except for very simple server naming support in the keystore structures.

type ServerInfo struct {
	Name    string
	Address string
	Port    string
	Key     *security.KeyInfo
}

func NewServerInfo(name, address, port string, key *security.KeyInfo) *ServerInfo {
	return &ServerInfo{
		Name:    name,
		Address: address,
		Port:    port,
		Key:     key.Clone(),
	}
}

func (si *ServerInfo) Clone() *ServerInfo {
	var clonedKey *security.KeyInfo
	if si.Key != nil {
		clonedKey = si.Key.Clone()
	}

	return &ServerInfo{
		Name:    si.Name,
		Address: si.Address,
		Port:    si.Port,
		Key:     clonedKey,
	}
}

func (sks *SimpleKeyStore) SetServerInfo(name, address, port string, key *security.KeyInfo) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	sks.Server = NewServerInfo(name, address, port, key)
	sks.Details.IsDirty = true
}

func (sks *SimpleKeyStore) RemoveServerInfo() {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	sks.Server = nil
	sks.Details.IsDirty = true
}

func (sks *SimpleKeyStore) UpdateServerKey(name string, key *security.KeyInfo) (found bool) {
	sks.SyncStore.Lock()
	defer sks.SyncStore.Unlock()

	if sks.Server == nil {
		return false
	}

	sks.Server.Key = key.Clone()
	sks.Details.IsDirty = true
	return true
}

func (sks *SimpleKeyStore) GetServerInfo() *ServerInfo {
	sks.SyncStore.RLock()
	sks.SyncStore.RUnlock()

	return sks.Server.Clone()
}
