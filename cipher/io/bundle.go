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
	cryptorand "crypto/rand"
	"github.com/thoughtrealm/bumblebee/security"
	"time"
)

const (
	BundleHeaderVersion = "1"
	BundleDataVersion   = "1"
	CHUNK_SIZE          = 32000
)

type BundleInputSource int

const (
	BundleInputSourceDirect BundleInputSource = iota
	BundleInputSourceFile
)

func BundleInputSourceToText(bis BundleInputSource) string {
	switch bis {
	case BundleInputSourceDirect:
		return "Direct"
	case BundleInputSourceFile:
		return "File"
	default:
		return "Unknown"
	}
}

type BundleInfo struct {
	// SymmetricKey is a random value used to encrypt the payload using Chacha20/Poly1305
	SymmetricKey []byte
	// Salt is a random value provided for the payload encryption
	Salt []byte
	// InputSource records the source type of the data provided for bundling
	InputSource BundleInputSource
	// The date the bundle was created
	CreateDate string // RFC3339
	// OriginalFileName records the file name of the source file, IF the source was a file
	OriginalFileName string
	// OriginalFileData records the date stamp of the source file, IF the source was a file
	OriginalFileDate string // RFC3339
	// ToName indicates the name used to identity the User public keys in the keystore
	ToName string
	// FromName indicates the name used to identity the keypair set that encrypted the bundle
	FromName string
	// SenderSig contains the RandomSignatureData struct data
	SenderSig []byte
	// HdrVer identifies the version of the bee functionality that built the hdr
	HdrVer string
	// PayloadVer identifies the version of the bee functionality that builtthe payload
	PayloadVer string
}

// NewBundle returns a BundleInfo that is pre-populated with a random symmetric key
func NewBundle() (*BundleInfo, error) {
	const SALT_SIZE = 64

	// We only set the "create date" value.  The caller must set other fields as relevant.
	newBundle := &BundleInfo{
		CreateDate: time.Now().Format(time.RFC3339),
		HdrVer:     BundleHeaderVersion,
		PayloadVer: BundleDataVersion,
	}

	// Generate random key... this will be strengthened and salted using Argon2
	newBundle.SymmetricKey = make([]byte, SALT_SIZE)
	_, err := cryptorand.Read(newBundle.SymmetricKey)
	if err != nil {
		return nil, err
	}

	return newBundle, nil
}

func (bundle *BundleInfo) Wipe() {
	if len(bundle.SymmetricKey) != 0 {
		security.Wipe(bundle.SymmetricKey)
	}

	if len(bundle.Salt) != 0 {
		security.Wipe(bundle.Salt)
	}

	if len(bundle.SenderSig) != 0 {
		security.Wipe(bundle.SenderSig)
	}
}
