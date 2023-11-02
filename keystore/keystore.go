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
Package keystore contains the functionality for the Bumblebee keystore.  It provides mechanisms
for storing key constructions relating to users, servers, groups, etc.

This package is just the functionality for the keystore itself.  A network service for providing
remote access will exist in a separate package.

The Bumblebee keystore has the following goals...
  - Support symmetric keys and NATS Keypair constructions.
  - Support for storing locally in a file system, generally assumed in the user's private file space.
  - Support for remote access by means of a separate network service of some type,
    likely some type of configurable HTTP/WebService/API.
  - Support for encrypted and unencrypted storage forms?
  - Support viewing keystore data through the bee app.
  - Support multi-file storage if grows large.
  - Support a concept of source to know where the item originated, like "office" vs local.  Should be able to
    have two with similar names but different source.
  - Items should have a unique identifier, maybe UUID, to use for syncing and precise identification.
  - Support for backups and restores from the bee app.
  - The remote functionality should support...
    -- Support a concept of users with properties: Name, Email, Password, Public Key, IsAdmin
    -- Support change logs and/or access logs for auditing
    -- Support include and exclude lists
    -- Maybe the ability to add user lists via text file import of a CSV or whatever
*/
package keystore

import "github.com/thoughtrealm/bumblebee/security"

var GlobalKeyStore KeyStore

type KeyStore interface {
	AddKey(name string, publickey []byte) error
	Count() int
	GetKey(name string) *security.Entity
	RenameEntity(oldName, newName string) (bool, error)
	RemoveEntity(name string) (found bool, err error)
	UpdatePublicKey(name string, publicKey string) (found bool, err error)
	Walk(info *WalkInfo) error
	WalkCount(nameMatchFilter string, walkFilterFunc KeyStoreWalkFilterFunc) (count int, err error)
	WriteToFile(filePath string) error
}

type KeyStoreWalkFunc func(entity *security.Entity)
type KeyStoreWalkFilterFunc func(entity *security.Entity) bool

type WalkInfo struct {
	NameMatchFilter string
	SortResults     bool
	WalkFilterFunc  KeyStoreWalkFilterFunc
	WalkFunc        KeyStoreWalkFunc
}

func NewWalkInfo(
	nameMatchFilter string,
	sortResults bool,
	walkFilterFunc KeyStoreWalkFilterFunc,
	walkFunc KeyStoreWalkFunc) *WalkInfo {

	return &WalkInfo{
		NameMatchFilter: nameMatchFilter,
		SortResults:     sortResults,
		WalkFilterFunc:  walkFilterFunc,
		WalkFunc:        walkFunc,
	}
}

func New(name, owner string, isLocal bool) KeyStore {
	return newSimpleKeyStore(name, owner, isLocal)
}

// NewFromMemory returns a new keystore from the byte sequence.
// The byte sequence would be bytes read from a keystore file, etc.
func NewFromMemory(bytesStore []byte) (newKeyStore KeyStore, err error) {
	return newSimpleKeyStoreFromMemory(bytesStore)
}

// NewFromFile returns a new keystore read from a file.
// The key sequence is used if the file is encrypted.  If it is not encrypted, the key is ignored.
func NewFromFile(storeKey []byte, filePath string) (newKeyStore KeyStore, err error) {
	return newSimpleKeyStoreFromFile(filePath)
}
