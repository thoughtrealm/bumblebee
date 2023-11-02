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
	"log"
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

/*
func (s *KeyStoreIOTestSuite) TestSimplexKeyStore_ReadWriteMemoryCycle_OneEntity_MsgPack() {
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
	s.Assert().Equal(s.testStore.Details.Name, newStore.Details.Name)
	s.Assert().Equal(s.testStore.Details.Owner, newStore.Details.Owner)
	s.Assert().Equal(s.testStore.Details.IsLocal, newStore.Details.IsLocal)

	// Compare store server info
	s.Assert().Equal(s.testStore.Server.Name, newStore.Server.Name)
	s.Assert().Equal(s.testStore.Server.Address, newStore.Server.Address)
	s.Assert().Equal(s.testStore.Server.Port, newStore.Server.Port)
	s.Assert().Equal(s.testStore.Server.Key.Name, newStore.Server.Key.Name)
	s.Assert().Equal(s.testStore.Server.Key.KeyType, newStore.Server.Key.KeyType)
	s.Assert().Equal(s.testStore.Server.Key.IsDefault, newStore.Server.Key.IsDefault)
	s.Assert().Equal(s.testStore.Server.Key.KeyData, newStore.Server.Key.KeyData)

	// Compare store entities
	s.Assert().Equal(len(s.testStore.Entities), len(newStore.Entities))
	for _, entity := range s.testStore.Entities {
		newEntity := newStore.GetEntity(true, entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.Key.Name, newEntity.Key.Name)
		s.Assert().Equal(entity.Key.KeyType, newEntity.Key.KeyType)
		s.Assert().Equal(entity.Key.IsDefault, newEntity.Key.IsDefault)
		s.Assert().Equal(entity.Key.KeyData, newEntity.Key.KeyData)
	}
}
*/

/*
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
	s.Assert().Equal(testStore.Details.Name, newStore.Details.Name)
	s.Assert().Equal(testStore.Details.Owner, newStore.Details.Owner)
	s.Assert().Equal(testStore.Details.IsLocal, newStore.Details.IsLocal)

	// Compare store server info
	s.Assert().Equal(testStore.Server.Name, newStore.Server.Name)
	s.Assert().Equal(testStore.Server.Address, newStore.Server.Address)
	s.Assert().Equal(testStore.Server.Port, newStore.Server.Port)
	s.Assert().Equal(testStore.Server.Key.Name, newStore.Server.Key.Name)
	s.Assert().Equal(testStore.Server.Key.KeyType, newStore.Server.Key.KeyType)
	s.Assert().Equal(testStore.Server.Key.IsDefault, newStore.Server.Key.IsDefault)
	s.Assert().Equal(testStore.Server.Key.KeyData, newStore.Server.Key.KeyData)

	// Compare store entities
	s.Assert().Equal(len(testStore.Entities), len(newStore.Entities))
	for _, entity := range testStore.Entities {
		newEntity := newStore.GetEntity(true, entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.Key.Name, newEntity.Key.Name)
		s.Assert().Equal(entity.Key.KeyType, newEntity.Key.KeyType)
		s.Assert().Equal(entity.Key.IsDefault, newEntity.Key.IsDefault)
		s.Assert().Equal(entity.Key.KeyData, newEntity.Key.KeyData)
	}
}
*/

/*
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
	s.Assert().Equal(testStore.Details.Name, newStore.Details.Name)
	s.Assert().Equal(testStore.Details.Owner, newStore.Details.Owner)
	s.Assert().Equal(testStore.Details.IsLocal, newStore.Details.IsLocal)

	// Compare store server info
	s.Assert().Equal(testStore.Server.Name, newStore.Server.Name)
	s.Assert().Equal(testStore.Server.Address, newStore.Server.Address)
	s.Assert().Equal(testStore.Server.Port, newStore.Server.Port)
	s.Assert().Equal(testStore.Server.Key.Name, newStore.Server.Key.Name)
	s.Assert().Equal(testStore.Server.Key.KeyType, newStore.Server.Key.KeyType)
	s.Assert().Equal(testStore.Server.Key.IsDefault, newStore.Server.Key.IsDefault)
	s.Assert().Equal(testStore.Server.Key.KeyData, newStore.Server.Key.KeyData)

	// Compare store entities
	s.Assert().Equal(len(testStore.Entities), len(newStore.Entities))
	for _, entity := range testStore.Entities {
		newEntity := newStore.GetEntity(true, entity.Name)
		s.Assert().Equal(entity.Name, newEntity.Name)
		s.Assert().Equal(entity.Key.Name, newEntity.Key.Name)
		s.Assert().Equal(entity.Key.KeyType, newEntity.Key.KeyType)
		s.Assert().Equal(entity.Key.IsDefault, newEntity.Key.IsDefault)
		s.Assert().Equal(entity.Key.KeyData, newEntity.Key.KeyData)
	}
}
*/

func BenchmarkWriteToMemoryMsgPack(b *testing.B) {
	entityCounts := []struct {
		entities int
	}{
		{1},
		{100},
		{1000},
		{5000},
		{10000},
		{100000},
		{1000000},
	}

	for idx, entityCount := range entityCounts {
		testStore := buildTestStoreMultiEntity(entityCount.entities)

		b.Run(fmt.Sprintf("[%d of %d] WriteToMemory_StoreOf_%d_Entities", idx+1, len(entityCounts), entityCount.entities), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = testStore.WriteToMemory()
			}
		})
	}
}

func BenchmarkReadFromMemoryWithMsgPack(b *testing.B) {
	entityCounts := []struct {
		entities int
	}{
		{1},
		{100},
		{1000},
		{5000},
		{10000},
		{100000},
		{1000000},
	}

	for idx, entityCount := range entityCounts {
		testStore := buildTestStoreMultiEntity(entityCount.entities)
		bytesStore, _ := testStore.WriteToMemory()

		b.Run(fmt.Sprintf("[%d of %d] ReadFromMemory_StoreOf_%d_Entities", idx+1, len(entityCounts), entityCount.entities), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := NewFromMemory(bytesStore)
				if err != nil {
					b.Fatalf("failed create new store: %s", err)
				}
			}
		})
	}
}

func BenchmarkGetEntity_OneEntity_NoLock(b *testing.B) {
	const ENTITIES = 1

	log.Printf("Building store with %d entities", ENTITIES)
	testStore := buildTestStoreMultiEntity(1000000)

	log.Println("Building prebuilt search key array")
	prebuiltSearchKeys := [ENTITIES]string{}
	for count := 0; count < ENTITIES; count++ {
		prebuiltSearchKeys[count] = fmt.Sprintf("BOB-%d", count)
	}

	log.Println("Running test")
	searchInt := 0
	b.Run(fmt.Sprintf("GetEntity_StoreOf_%d_Entities", ENTITIES), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := testStore.GetEntity(prebuiltSearchKeys[searchInt])
			if entity == nil {
				b.Fatalf("No entity found")
			}
			searchInt += 1
			if searchInt >= ENTITIES {
				searchInt = 0
			}
		}
	})
}

func BenchmarkGetEntity_1000Entities_NoLock(b *testing.B) {
	const ENTITIES = 1000

	log.Printf("Building store with %d entities", ENTITIES)
	testStore := buildTestStoreMultiEntity(1000000)

	log.Println("Building prebuilt search key array")
	prebuiltSearchKeys := [ENTITIES]string{}
	for count := 0; count < ENTITIES; count++ {
		prebuiltSearchKeys[count] = fmt.Sprintf("BOB-%d", count)
	}

	log.Println("Running test")
	searchInt := 0
	b.Run(fmt.Sprintf("GetEntity_StoreOf_%d_Entities", ENTITIES), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := testStore.GetEntity(prebuiltSearchKeys[searchInt])
			if entity == nil {
				b.Fatalf("No entity found")
			}
			searchInt += 1
			if searchInt >= ENTITIES {
				searchInt = 0
			}
		}
	})
}

func BenchmarkGetEntity_1000000Entities_NoLock(b *testing.B) {
	const ENTITIES = 1000000

	log.Printf("Building store with %d entities", ENTITIES)
	testStore := buildTestStoreMultiEntity(1000000)

	log.Println("Building prebuilt search key array")
	prebuiltSearchKeys := [ENTITIES]string{}
	for count := 0; count < ENTITIES; count++ {
		prebuiltSearchKeys[count] = fmt.Sprintf("BOB-%d", count)
	}

	log.Println("Running test")
	searchInt := 0
	b.Run(fmt.Sprintf("GetEntity_StoreOf_%d_Entities", ENTITIES), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := testStore.GetEntity(prebuiltSearchKeys[searchInt])
			if entity == nil {
				b.Fatalf("No entity found")
			}
			searchInt += 1
			if searchInt >= ENTITIES {
				searchInt = 0
			}
		}
	})
}
