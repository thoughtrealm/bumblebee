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

package helpers

import cryptorand "crypto/rand"

func GetRandomBytes(requestedByteCount int) (secretBytes []byte, err error) {
	const blockSize = 1000

	if requestedByteCount == 0 {
		return nil, nil
	}

	secretBytes = make([]byte, 0, requestedByteCount)

	// We read random sets in smaller block sizes and aggregate into the secret bytes
	bytesRead := 0
	bytesToRead := 0
	for {
		bytesToRead = requestedByteCount - bytesRead
		if bytesToRead == 0 {
			// not sure why this would happen, but just to prevent endless looping for some odd, unforeseen scenario
			return secretBytes, nil
		}

		if bytesToRead > blockSize {
			bytesToRead = blockSize
		}

		blockBytes := make([]byte, bytesToRead)
		_, err = cryptorand.Read(blockBytes)
		if err != nil {
			return nil, err
		}

		secretBytes = append(secretBytes, blockBytes...)
		bytesRead += bytesToRead
		if bytesRead >= requestedByteCount {
			return secretBytes, nil
		}
	}
}
