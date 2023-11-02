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

import "errors"

func IntToUint8Bytes(valueInt int) (valueBytes []byte) {
	valueBytes = make([]byte, 1)
	valueUint8 := uint8(valueInt)
	valueBytes[0] = valueUint8
	return
}

func Uint8BytesToInt(valueBytes []byte) (valueInt int, err error) {
	if len(valueBytes) != 1 {
		return 0, errors.New("input bytes invalid size")
	}

	valueUint8 := uint8(valueBytes[0])
	return int(valueUint8), nil
}

func IntToUint16Bytes(valueInt int) (valueBytes []byte) {
	valueBytes = make([]byte, 2)
	valueUint16 := uint16(valueInt)
	valueBytes[0] = byte((valueUint16 & 0xFF00) >> 8)
	valueBytes[1] = byte(valueUint16 & 0x00FF)
	return
}

func Uint16BytesToInt(valueBytes []byte) (valueInt int, err error) {
	if len(valueBytes) != 2 {
		return 0, errors.New("input bytes invalid size")
	}

	valueUint16 := (uint16(valueBytes[0]) << 8) + uint16(valueBytes[1])
	return int(valueUint16), nil
}
