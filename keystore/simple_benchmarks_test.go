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
	"log"
	"testing"
)

func BenchmarkWriteToMemory(b *testing.B) {
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

func BenchmarkReadFromMemory(b *testing.B) {
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
