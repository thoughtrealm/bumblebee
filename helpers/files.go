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

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ReplaceFileExt will change the file extension if the filePath contains one.
// If it does not, it will add the extension.  newExt is allowed to have a period
// or not, both scenarios will work correctly
func ReplaceFileExt(filePath, newExt string) string {
	if newExt == "" {
		return filePath
	}

	var newExtPeriod string
	if !strings.HasPrefix(newExt, ".") {
		newExtPeriod = "."
	}

	ext := filepath.Ext(filePath)
	if ext == "" {
		return filePath + newExtPeriod + newExt
	}

	return filePath[:len(filePath)-len(ext)] + newExtPeriod + newExt
}

func FileExistsWithDetails(filePath string) (isFound, isDir bool, err error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, false, nil
		}

		return false, false, err

	}

	return true, info.IsDir(), nil
}

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}

func IncludeTrailingPathSeparator(aPath string) string {
	if strings.HasSuffix(aPath, string(filepath.Separator)) {
		return aPath
	}

	return aPath + string(filepath.Separator)
}

func RemoveTrailingPathSeparator(aPath string) string {
	if !strings.HasSuffix(aPath, string(filepath.Separator)) {
		return aPath
	}

	return aPath[:len(aPath)-1]
}

func PathExistsInfo(filePath string) (found, isDir bool) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, false
	}

	if info.IsDir() {
		return true, true
	}

	return true, false
}

// GetFileSafeName will replace characters in the inputName that are not
// safe for naming directories or files.  Due to cross-platform concerns,
// this will convert or remove things that are not within the POSIX Portable File Name
// character set.
func GetFileSafeName(inputName string) (outputName string) {
	const POSIX_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"._-"

	for i := 0; i < len(inputName); i++ {
		if inputName[i] == ' ' {
			outputName += "-"
			continue
		}

		if !strings.Contains(POSIX_CHARS, string(inputName[i])) {
			outputName += "_"
			continue
		}

		outputName += string(inputName[i])
	}

	return outputName
}

func GetEnvSafeName(inputName string) (outputName string) {
	const POSIX_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"._"

	for i := 0; i < len(inputName); i++ {
		if inputName[i] == ' ' {
			outputName += "_"
			continue
		}

		if !strings.Contains(POSIX_CHARS, string(inputName[i])) {
			outputName += "_"
			continue
		}

		outputName += string(inputName[i])
	}

	return outputName
}

func ForcePath(p string) error {
	return os.MkdirAll(p, os.ModePerm)
}
