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

package keypairs

import (
	"github.com/stretchr/testify/assert"
	"github.com/thoughtrealm/bumblebee/helpers"
	"os"
	"path/filepath"
	"testing"
)

func TestNewKeypairStore(t *testing.T) {
	newKPStore := NewKeypairStore()
	assert.NotNil(t, newKPStore)
}

func TestNewKeypairStoreWithKeypair(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, kpi) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	kpiStore := newKPStore.GetKeyPairInfo("test")
	if !assert.NotNil(t, kpiStore) {
		return
	}

	assert.Equal(t, "test", kpiStore.Name)
}

func TestNewKeypairStoreFromFile(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, "test", kpi.Name) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	if !assert.Nil(t, helpers.ForcePath("testfiles")) {
		// can't test if we can't create the output file path
		return
	}
	defer func() {
		_ = os.RemoveAll("testfiles")
	}()

	storePath := filepath.Join("testfiles", "teststore")
	err = newKPStore.SaveKeyPairStore(nil, storePath)
	if !assert.Nil(t, err) {
		return
	}

	storeFromFile, err := NewKeypairStoreFromFile(nil, storePath)
	if !assert.Nil(t, err) {
		return
	}

	if !assert.NotNil(t, storeFromFile) {
		return
	}

	kpiFile := newKPStore.GetKeyPairInfo("test")
	if !assert.NotNil(t, kpiFile) {
		return
	}

	assert.Equal(t, "test", kpiFile.Name)
}

func TestNewKeypairStoreFromFile_RequiresPassword_FailsNoPassword(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, "test", kpi.Name) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	if !assert.Nil(t, helpers.ForcePath("testfiles")) {
		// can't test if we can't create the output file path
		return
	}
	defer func() {
		_ = os.RemoveAll("testfiles")
	}()

	storePath := filepath.Join("testfiles", "teststore")
	err = newKPStore.SaveKeyPairStore([]byte("password"), storePath)
	if !assert.Nil(t, err) {
		return
	}

	// Not providing a password but one is required
	_, err = NewKeypairStoreFromFile(nil, storePath)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "decrypt failed")
}

func TestNewKeypairStoreFromFile_RequiresPassword_FailsWrongPassword(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, "test", kpi.Name) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	if !assert.Nil(t, helpers.ForcePath("testfiles")) {
		// can't test if we can't create the output file path
		return
	}
	defer func() {
		_ = os.RemoveAll("testfiles")
	}()

	storePath := filepath.Join("testfiles", "teststore")
	err = newKPStore.SaveKeyPairStore([]byte("password"), storePath)
	if !assert.Nil(t, err) {
		return
	}

	// Providing wrong password
	_, err = NewKeypairStoreFromFile([]byte("wrongpassword"), storePath)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "decrypt failed")
}

func TestNewKeypairStoreFromFile_RequiresNoPassword_PassesWithIgnoredPassword(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, "test", kpi.Name) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	if !assert.Nil(t, helpers.ForcePath("testfiles")) {
		// can't test if we can't create the output file path
		return
	}
	defer func() {
		_ = os.RemoveAll("testfiles")
	}()

	storePath := filepath.Join("testfiles", "teststore")
	err = newKPStore.SaveKeyPairStore(nil, storePath)
	if !assert.Nil(t, err) {
		return
	}

	// Providing password when one is not required, but it gets ignored,
	// because code detects it is not password encoded and ignores the passed in password.
	_, err = NewKeypairStoreFromFile([]byte("password"), storePath)
	assert.Nil(t, err)
}

func TestNewKeypairStoreFromFile_RequiresPassword_PassesWithCorrectPassword(t *testing.T) {
	newKPStore, kpi, err := NewKeypairStoreWithKeypair("test")
	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, "test", kpi.Name) {
		return
	}

	if !assert.NotNil(t, newKPStore) {
		return
	}

	if !assert.Nil(t, helpers.ForcePath("testfiles")) {
		// can't test if we can't create the output file path
		return
	}
	defer func() {
		_ = os.RemoveAll("testfiles")
	}()

	storePath := filepath.Join("testfiles", "teststore")
	err = newKPStore.SaveKeyPairStore([]byte("password"), storePath)
	if !assert.Nil(t, err) {
		return
	}

	// Providing wrong password
	_, err = NewKeypairStoreFromFile([]byte("password"), storePath)
	assert.Nil(t, err)
}
