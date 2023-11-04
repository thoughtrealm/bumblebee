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

package main

import (
	"flag"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Expect pattern:
// distributions count datasize

var (
	testCount    int
	testMsecs    int
	testProcs    int
	testDataSize int

	totalBytesWritten int64
)

func main() {
	var (
		err error
	)

	flag.IntVar(&testCount, "count", 0, "The count of iterations for distribution analsysis")
	flag.IntVar(&testMsecs, "msecs", 1000, "The msecs to run the timed test")
	flag.IntVar(&testProcs, "procs", 1, "The number of procs to use for the timed test")
	flag.IntVar(&testDataSize, "datasize", 1000, "The size in bytes to encrypt in the timed test")

	flag.Parse()

	fmt.Println("Distributions analysis for Bumblebee constructions")
	fmt.Println("")

	p := message.NewPrinter(language.English)

	fmt.Println("Using these input values...")
	_, _ = p.Printf("  - Test count     : %d\n", testCount)
	_, _ = p.Printf("  - Test msecs     : %d\n", testMsecs)
	_, _ = p.Printf("  - Test procs     : %d\n", testProcs)
	_, _ = p.Printf("  - Test data size : %d\n", testDataSize)

	fmt.Println("")
	fmt.Println("Initializing timed test mgr...")

	mgr := newTimedTestMgr(testProcs, testMsecs, testDataSize)
	mgr.init()

	fmt.Println("Starting timed test...")

	err = mgr.runMgr()
	if err != nil {
		fmt.Printf("Error running test: %s\n", err)
		return
	}

	fmt.Println("")
	mgr.printStats()

	fmt.Println("\nTimed test complete")
	fmt.Println("")

	fmt.Println("** Distribution analysis not yet implemented **")
}
