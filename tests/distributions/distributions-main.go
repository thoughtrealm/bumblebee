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
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"strconv"
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

	defer fmt.Println("\nDistributions analysis complete")

	fmt.Println("Distributions analysis for Bumblebee constructions")
	fmt.Println("")

	// os.Args should be 3... 1 for the os app ref, then 2 for our arguments

	if len(os.Args) != 5 {
		fmt.Printf("Error: Expected three arguments.  Got %d arguments\n", len(os.Args))
		fmt.Println("Input format: \"distributions <count> <msecs> <procs> <data-size>\"")
		return
	}

	testCount, err = strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Test count value \"%s\" is not a valid integer\n", os.Args[1])
		return
	}

	testMsecs, err = strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Test msecs value \"%s\" is not a valid integer\n", os.Args[2])
		return
	}

	testProcs, err = strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("Test procs value \"%s\" is not a valid integer\n", os.Args[3])
		return
	}

	testDataSize, err = strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Printf("Test data size value \"%s\" is not a valid integer\n", os.Args[4])
		return
	}

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
}
