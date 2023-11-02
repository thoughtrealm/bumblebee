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
	"time"
)

type procMessage int

const (
	procMessageDone procMessage = iota
	procMessageErr
)

type procMessageInfo struct {
	msg          procMessage
	id           int
	bytesWritten int
	err          error
	testDuration time.Duration
}

type procInfo struct {
	testProc       *timedTestProc
	id             int
	count          int
	errCount       int
	totalDuration  time.Duration
	resultsPending bool
}

type timedTestMgr struct {
	procCount  int
	timedMsecs int
	dataSize   int ``
	chanMsg    chan procMessageInfo
	procs      []*procInfo
	startTime  time.Time
	endTime    time.Time
}

func newTimedTestMgr(procCount, testMsecs, dataSize int) *timedTestMgr {
	newMgr := &timedTestMgr{
		procCount:  procCount,
		timedMsecs: testMsecs,
		dataSize:   dataSize,
		chanMsg:    make(chan procMessageInfo, 10),
	}

	return newMgr
}

func (ttm *timedTestMgr) init() {
	for i := 0; i < ttm.procCount; i++ {
		testProc := newTimedTestProc(i+1, ttm)
		testProc.init(ttm.dataSize)
		ttm.procs = append(ttm.procs, &procInfo{
			testProc:      testProc,
			id:            i + 1,
			count:         0,
			totalDuration: 0,
		})

	}
}

// runMgr will block until the test is done.  Might change that in the future.
func (ttm *timedTestMgr) runMgr() error {
	timer := time.NewTimer(time.Duration(ttm.timedMsecs) * time.Millisecond)
	defer timer.Stop()

	chanTimerWait := make(chan struct{}, 1)
	chanCancelTimerWait := make(chan struct{}, 1)

	// we use a go proc to inform us of the timer status, so we don't have to exit the master select after the timer fires
	go func() {
		select {
		case <-timer.C:
			close(chanTimerWait)
			break
		case <-chanCancelTimerWait:
			break
		}
	}()

	defer close(chanCancelTimerWait)

	ttm.startTime = time.Now()
	go ttm.startTest()
	defer func() {
		ttm.endTime = time.Now()
	}()

	testComplete := false
	shouldEnd := false
	for {
		select {
		case msg := <-ttm.chanMsg:
			ttm.handle(msg, testComplete)
			if testComplete && ttm.allProcsDone() {
				shouldEnd = true
			}
		case <-chanTimerWait:
			chanTimerWait = nil

			// setting testComplete indicates to exit the mgr once the remaining tests are completed
			testComplete = true
			if ttm.allProcsDone() {
				shouldEnd = true
			}
		}

		if shouldEnd {
			break
		}
	}

	return nil
}

func (ttm *timedTestMgr) allProcsDone() bool {
	for _, testProc := range ttm.procs {
		if testProc.resultsPending {
			return false
		}
	}

	return true
}

func (ttm *timedTestMgr) startTest() {
	for _, ttp := range ttm.procs {
		ttp.resultsPending = true
		ttp.testProc.startTest()
	}
}

func (ttm *timedTestMgr) handle(msg procMessageInfo, testComplete bool) {
	switch msg.msg {
	case procMessageDone:
		ttm.handleDone(msg, testComplete)
	case procMessageErr:
		ttm.handleErr(msg, testComplete)
	}
}

func (ttm *timedTestMgr) handleDone(msg procMessageInfo, testComplete bool) {
	pi := ttm.findProcById(msg.id)
	if pi == nil {
		panic(fmt.Sprintf("Unknown testProc id %d in handleDone()", msg.id))
	}

	pi.count += 1
	pi.resultsPending = false
	pi.totalDuration += msg.testDuration

	if !testComplete {
		pi.resultsPending = true
		pi.testProc.startTest()
	}
}

func (ttm *timedTestMgr) handleErr(msg procMessageInfo, testComplete bool) {
	pi := ttm.findProcById(msg.id)
	if pi == nil {
		panic(fmt.Sprintf("Unknown testProc id %d in handleErr()", msg.id))
	}

	pi.errCount += 1
	pi.resultsPending = false

	// Todo: for now, if a testProc returns error, we DO restart it...is that good or bad?
	if !testComplete {
		pi.resultsPending = true
		pi.testProc.startTest()
	}
}

func (ttm *timedTestMgr) findProcById(id int) *procInfo {
	for _, pi := range ttm.procs {
		if pi.id == id {
			return pi
		}
	}

	return nil
}

func (ttm *timedTestMgr) printStats() {
	fmt.Println("Test Proc Stats")
	fmt.Println("=======================================")
	var (
		grandTotalDuration time.Duration
		totalCompleted     int
		totalErrors        int
	)
	for idx, pi := range ttm.procs {
		grandTotalDuration += pi.totalDuration
		totalCompleted += pi.count
		totalErrors += pi.errCount

		fmt.Printf("Proc %02d of %d...\n", idx+1, len(ttm.procs))
		fmt.Printf("  - Tests Completed : %d\n", pi.count)
		fmt.Printf("  - Tests Failed    : %d\n", pi.errCount)
		fmt.Printf("  - Avg time        : %d ms\n", (pi.totalDuration/time.Duration(pi.count))/time.Millisecond)
		fmt.Println("")
	}

	fmt.Println("")
	fmt.Println("Total Test Stats")
	fmt.Println("=======================================")
	fmt.Printf("Test Start Time    : %s\n", ttm.startTime.Format("2006-01-02 15:04:05.000000000"))
	fmt.Printf("Test End Time      : %s\n", ttm.endTime.Format("2006-01-02 15:04:05.000000000"))
	fmt.Printf("Total Completed    : %d\n", totalCompleted)
	fmt.Printf("Total Errors       : %d\n", totalErrors)
	fmt.Printf("Avg per second:    : %d\n", time.Duration(totalCompleted)/(ttm.endTime.Sub(ttm.startTime)/time.Second))
	fmt.Printf("Avg time all procs : %d ms", (grandTotalDuration/time.Millisecond)/time.Duration(totalCompleted))
	fmt.Println("")
}
