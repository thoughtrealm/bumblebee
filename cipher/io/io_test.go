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

package io

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/thoughtrealm/bumblebee/helpers"
	"github.com/thoughtrealm/bumblebee/security"
	"os"
	"path/filepath"
	"testing"
)

const test_path = "testfiles"
const werner_text = "My name is Werner Brandon.  My voice is my passport.  Verify me."

var werner_bytes = []byte(werner_text)

type CipherIOTestSuite struct {
	suite.Suite
	test_path string
}

func TestCipherIOTestSuite(t *testing.T) {
	suite.Run(t, new(CipherIOTestSuite))
}

func (s *CipherIOTestSuite) SetupSuite() {
	s.test_path = "testfiles"

	if err := helpers.ForcePath(test_path); err != nil {
		s.T().Logf("Unable to create output path \"%s\"", test_path)
		return
	}
}

func (s *CipherIOTestSuite) TearDownSuite() {
	_ = os.RemoveAll(test_path)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToCombinedFileFromReader() {
	var encryptedFile = filepath.Join(test_path, "test1.combined")
	var decryptedFile = filepath.Join(test_path, "test1.decrypted")

	defer func() {
		if helpers.FileExists(encryptedFile) {
			_ = os.Remove(encryptedFile)
		}
	}()

	defer func() {
		if helpers.FileExists(decryptedFile) {
			_ = os.Remove(decryptedFile)
		}
	}()

	secretBytes := werner_bytes
	readerBuff := bytes.NewBuffer(secretBytes)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToCombinedFileFromReader(encryptedFile, readerBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the file
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfr.ReadCombinedFileToFile(encryptedFile, decryptedFile)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBytes, err := os.ReadFile(decryptedFile)
	if !assert.Nil(s.T(), err) {
		return
	}

	assert.Equal(s.T(), secretBytes, decryptedBytes)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToSplitFilesFromReader() {
	var encryptedFileHdr = filepath.Join(test_path, "test1.bhdr")
	var encryptedFileData = filepath.Join(test_path, "test1.bdata")
	var decryptedFile = filepath.Join(test_path, "test1.decrypted")

	defer func() {
		if helpers.FileExists(encryptedFileHdr) {
			_ = os.Remove(encryptedFileHdr)
		}
	}()

	defer func() {
		if helpers.FileExists(encryptedFileData) {
			_ = os.Remove(encryptedFileData)
		}
	}()

	defer func() {
		if helpers.FileExists(decryptedFile) {
			_ = os.Remove(decryptedFile)
		}
	}()

	secretBytes := werner_bytes
	readerBuff := bytes.NewBuffer(secretBytes)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToSplitFilesFromReader(encryptedFileHdr, readerBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the file
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfr.ReadSplitFilesToFile(encryptedFileHdr, encryptedFileData, decryptedFile)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBytes, err := os.ReadFile(decryptedFile)
	if !assert.Nil(s.T(), err) {
		return
	}

	assert.Equal(s.T(), secretBytes, decryptedBytes)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToCombinedStreamFromReader() {
	secretBytes := werner_bytes
	readerBuff := bytes.NewBuffer(secretBytes)
	encryptedBuff := bytes.NewBuffer(nil)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToCombinedStreamFromReader(readerBuff, encryptedBuff, nil)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the buffer
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBuff := bytes.NewBuffer(nil)
	_, err = cfr.ReadCombinedStreamToWriter(encryptedBuff, decryptedBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	decryptedBytes := decryptedBuff.Bytes()
	assert.Equal(s.T(), secretBytes, decryptedBytes)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToCombinedStreamFromReader_LargeStream() {
	secretBytes, err := helpers.GetRandomBytes(10000000)

	readerBuff := bytes.NewBuffer(secretBytes)
	encryptedBuff := bytes.NewBuffer(nil)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToCombinedStreamFromReader(readerBuff, encryptedBuff, nil)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the buffer
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBuff := bytes.NewBuffer(nil)
	_, err = cfr.ReadCombinedStreamToWriter(encryptedBuff, decryptedBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	decryptedBytes := decryptedBuff.Bytes()
	assert.Equal(s.T(), secretBytes, decryptedBytes)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToSplitStreamsFromReader() {
	secretBytes := werner_bytes
	readerBuff := bytes.NewBuffer(secretBytes)
	encryptedBuffHdr := bytes.NewBuffer(nil)
	encryptedBuffData := bytes.NewBuffer(nil)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToSplitStreamsFromReader(readerBuff, encryptedBuffHdr, encryptedBuffData, nil, nil)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the buffer
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBuff := bytes.NewBuffer(nil)
	_, err = cfr.ReadSplitStreamsToWriter(encryptedBuffHdr, encryptedBuffData, decryptedBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	decryptedBytes := decryptedBuff.Bytes()
	assert.Equal(s.T(), secretBytes, decryptedBytes)
}

func (s *CipherIOTestSuite) TestCipherFileWriter_WriteToSplitStreamsFromReader_LargeStream() {
	secretBytes, err := helpers.GetRandomBytes(10000000)
	if !assert.Nil(s.T(), err) {
		return
	}

	readerBuff := bytes.NewBuffer(secretBytes)
	encryptedBuffHdr := bytes.NewBuffer(nil)
	encryptedBuffData := bytes.NewBuffer(nil)

	receiverKPI, _ := security.NewKeyPairInfoWithSeeds("receiverKPI")
	receiverCipherPublicKey, receiverSigningPublicKey, err := receiverKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	receiverKI, _ := security.NewKeyInfo("receiverKI", receiverCipherPublicKey, receiverSigningPublicKey)

	senderKPI, _ := security.NewKeyPairInfoWithSeeds("senderKPI")
	senderCipherPublicKey, senderSigningPublicKey, err := senderKPI.PublicKeys()
	if !assert.Nil(s.T(), err) {
		return
	}
	senderKI, _ := security.NewKeyInfo("senderKI", senderCipherPublicKey, senderSigningPublicKey)

	cfw, err := NewCipherWriter(receiverKI, senderKPI)
	if !assert.Nil(s.T(), err) {
		return
	}

	_, err = cfw.WriteToSplitStreamsFromReader(readerBuff, encryptedBuffHdr, encryptedBuffData, nil, nil)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to decrypt the buffer
	cfr, err := NewCipherFileReader(receiverKPI, senderKI)
	if !assert.Nil(s.T(), err) {
		return
	}

	// now, we want to compare the results
	decryptedBuff := bytes.NewBuffer(nil)
	_, err = cfr.ReadSplitStreamsToWriter(encryptedBuffHdr, encryptedBuffData, decryptedBuff)
	if !assert.Nil(s.T(), err) {
		return
	}

	decryptedBytes := decryptedBuff.Bytes()
	assert.Equal(s.T(), secretBytes, decryptedBytes)
}
