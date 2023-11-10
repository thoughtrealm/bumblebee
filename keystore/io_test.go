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

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
)

type KeyStoreIOTestSuite struct {
	suite.Suite
	testStore *SimpleKeyStore
}

func TestKeyStoreIOTestSuite(t *testing.T) {
	suite.Run(t, new(KeyStoreIOTestSuite))
}

func (s *KeyStoreIOTestSuite) SetupTest() {
	s.testStore = buildTestStore()
}

func (s *KeyStoreIOTestSuite) Test64BitLen() {
	byteSlice := make([]byte, 4000000000)
	s.Assert().Equal(4000000000, len(byteSlice))
}

func (s *KeyStoreIOTestSuite) TestSimplexKeyStore_ReadWriteMemoryCycle_OneEntity() {
	bytesStore, err := s.testStore.WriteToMemory()
	if !s.Assert().Nil(err) {
		return
	}
	s.Assert().NotNil(bytesStore)

	// Validate by reading it back into a test store structure
	newStore, err := NewFromMemory(bytesStore)
	if !s.Assert().Nil(err) {
		return
	}

	// Compare store details
	s.Assert().Equal(s.testStore.Details.Name, newStore.GetDetails().Name)
	s.Assert().Equal(s.testStore.Details.Owner, newStore.GetDetails().Owner)
	s.Assert().Equal(s.testStore.Details.IsLocal, newStore.GetDetails().IsLocal)

	// Compare store server info
	s.Assert().Equal(s.testStore.Server.Name, newStore.GetServerInfo().Name)
	s.Assert().Equal(s.testStore.Server.Address, newStore.GetServerInfo().Address)
	s.Assert().Equal(s.testStore.Server.Port, newStore.GetServerInfo().Port)
	s.Assert().Equal(s.testStore.Server.Key.Name, newStore.GetServerInfo().Key.Name)
	s.Assert().True(s.testStore.Server.Key.IsSameData(newStore.GetServerInfo().Key))

	// Compare store entities
	s.Assert().Equal(len(s.testStore.Entities), newStore.Count())
	for _, entity := range s.testStore.Entities {
		newEntity := newStore.GetKey(entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.PublicKeys.Name, newEntity.PublicKeys.Name)
		s.Assert().True(entity.PublicKeys.IsSameData(newEntity.PublicKeys))
	}
}

func (s *KeyStoreIOTestSuite) TestSimplexKeyStore_ReadWriteMemoryCycle_1000Entities() {
	testStore := buildTestStoreMultiEntity(1000)

	bytesStore, err := testStore.WriteToMemory()
	if !s.Assert().Nil(err) {
		return
	}
	s.Assert().NotNil(bytesStore)
	fmt.Printf("data len: %d\n", len(bytesStore))

	// Validate by reading it back into a test store structure
	newStore, err := NewFromMemory(bytesStore)
	if !s.Assert().Nil(err) {
		return
	}

	// Compare store details
	s.Assert().Equal(testStore.Details.Name, newStore.GetDetails().Name)
	s.Assert().Equal(testStore.Details.Owner, newStore.GetDetails().Owner)
	s.Assert().Equal(testStore.Details.IsLocal, newStore.GetDetails().IsLocal)

	// Compare store server info
	s.Assert().Equal(testStore.Server.Name, newStore.GetServerInfo().Name)
	s.Assert().Equal(testStore.Server.Address, newStore.GetServerInfo().Address)
	s.Assert().Equal(testStore.Server.Port, newStore.GetServerInfo().Port)
	s.Assert().Equal(testStore.Server.Key.Name, newStore.GetServerInfo().Key.Name)
	s.Assert().True(testStore.Server.Key.IsSameData(newStore.GetServerInfo().Key))

	// Compare store entities
	s.Assert().Equal(len(testStore.Entities), newStore.Count())
	for _, entity := range testStore.Entities {
		newEntity := newStore.GetKey(entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.PublicKeys.Name, newEntity.PublicKeys.Name)
		s.Assert().True(entity.PublicKeys.IsSameData(newEntity.PublicKeys))
	}
}

func (s *KeyStoreIOTestSuite) TestSimplexKeyStore_ReadWriteMemoryCycle_1000000Entities() {
	testStore := buildTestStoreMultiEntity(1000000)

	bytesStore, err := testStore.WriteToMemory()
	if !s.Assert().Nil(err) {
		return
	}
	s.Assert().NotNil(bytesStore)
	fmt.Printf("data len: %d", len(bytesStore))

	// Validate by reading it back into a test store structure
	newStore, err := NewFromMemory(bytesStore)
	if !s.Assert().Nil(err) {
		return
	}

	// Compare store details
	s.Assert().Equal(testStore.Details.Name, newStore.GetDetails().Name)
	s.Assert().Equal(testStore.Details.Owner, newStore.GetDetails().Owner)
	s.Assert().Equal(testStore.Details.IsLocal, newStore.GetDetails().IsLocal)

	// Compare store server info
	s.Assert().Equal(testStore.Server.Name, newStore.GetServerInfo().Name)
	s.Assert().Equal(testStore.Server.Address, newStore.GetServerInfo().Address)
	s.Assert().Equal(testStore.Server.Port, newStore.GetServerInfo().Port)
	s.Assert().Equal(testStore.Server.Key.Name, newStore.GetServerInfo().Key.Name)
	s.Assert().True(testStore.Server.Key.IsSameData(newStore.GetServerInfo().Key))

	// Compare store entities
	s.Assert().Equal(len(testStore.Entities), newStore.Count())
	for _, entity := range testStore.Entities {
		newEntity := newStore.GetKey(entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.PublicKeys.Name, newEntity.PublicKeys.Name)
		s.Assert().True(entity.PublicKeys.IsSameData(newEntity.PublicKeys))
	}
}
