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
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/nats-io/nkeys"
	cipherio "github.com/thoughtrealm/bumblebee/cipher/io"
	"github.com/thoughtrealm/bumblebee/security"
	"io"
	"time"
)

type timedTestProc struct {
	id           int
	readData     []byte
	readBuffer   *bytes.Reader
	outputBuffer *bytes.Buffer
	mgr          *timedTestMgr
	receiverKI   *security.KeyInfo
	senderKI     *security.KeyInfo
	cw           *cipherio.CipherWriter
}

func newTimedTestProc(id int, mgr *timedTestMgr) *timedTestProc {
	return &timedTestProc{id: id, mgr: mgr}
}

func (ttp *timedTestProc) init(dataSize int) {
	var err error

	ttp.readData = make([]byte, dataSize)
	_, err = rand.Read(ttp.readData)
	if err != nil {
		panic(fmt.Sprintf("error reading random bytes in ttp.init: %s", err))
	}

	ttp.readBuffer = bytes.NewReader(ttp.readData)

	err = ttp.createKeypairs()
	if err != nil {
		panic(fmt.Sprintf("error creating keypairs in ttp.init: %s", err))
	}

	ttp.cw, err = cipherio.NewCipherWriter(ttp.receiverKI, ttp.senderKI)
	if err != nil {
		panic(fmt.Sprintf("error creating new cipher writer in ttp.init: %s", err))
	}
}

func (ttp *timedTestProc) startTest() {
	go ttp.doTest()
}

func (ttp *timedTestProc) doTest() {
	startTime := time.Now()

	ttp.outputBuffer = bytes.NewBuffer(nil)
	bytesWritten, err := ttp.cw.WriteToCombinedStreamFromReader(ttp.readBuffer, ttp.outputBuffer, nil)
	if err != nil {
		ttp.mgr.chanMsg <- procMessageInfo{
			msg:          procMessageErr,
			id:           ttp.id,
			bytesWritten: bytesWritten,
			err:          fmt.Errorf("error during cipher write: %s", err),
			testDuration: time.Since(startTime),
		}
		return
	}

	_, err = ttp.readBuffer.Seek(0, io.SeekStart)
	if err != nil {
		ttp.mgr.chanMsg <- procMessageInfo{
			msg:          procMessageErr,
			id:           ttp.id,
			bytesWritten: bytesWritten,
			err:          fmt.Errorf("error resetting read buffer: %s", err),
			testDuration: time.Since(startTime),
		}
		return
	}

	ttp.mgr.chanMsg <- procMessageInfo{
		msg:          procMessageDone,
		id:           ttp.id,
		bytesWritten: bytesWritten,
		err:          nil,
		testDuration: time.Since(startTime),
	}
}

func (ttp *timedTestProc) createKeypairs() (err error) {
	var (
		receiverKP        nkeys.KeyPair
		receiverPublicKey string
		senderKP          nkeys.KeyPair
		senderSeed        []byte
	)

	receiverKP, err = nkeys.CreateCurveKeys()
	if err != nil {
		return fmt.Errorf("failed creating receiver kp: %s\n", err)
	}

	receiverPublicKey, err = receiverKP.PublicKey()
	if err != nil {
		return fmt.Errorf("failed creating reciever public key: %s", err)
	}

	ttp.receiverKI = &security.KeyInfo{
		IsDefault: false,
		KeyType:   security.KeyTypePublic,
		Name:      "receiver",
		KeyData:   []byte(receiverPublicKey),
	}

	senderKP, err = nkeys.CreateCurveKeys()
	if err != nil {
		return fmt.Errorf("failed creating sender kp: %s\n", err)
	}

	senderSeed, err = senderKP.Seed()
	if err != nil {
		return fmt.Errorf("failed creating sender private key: %s", err)
	}

	ttp.senderKI = &security.KeyInfo{
		IsDefault: false,
		KeyType:   security.KeyTypeSeed,
		Name:      "sender",
		KeyData:   senderSeed,
	}

	return nil
}
