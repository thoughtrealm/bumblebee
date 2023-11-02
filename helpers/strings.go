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
	"regexp"
	"strings"
)

// CompareStrings is a case-insensitive comparison.  Returns:
//   - 0 if strings are the same
//   - -1 if a is less than b
//   - 1 if a is greater than b
//
// While strings.Equalfold is available from the standard lib, it does not tell
// us which value is greater than the other.  This CompareStrings provides that where needed.
// When we only need a true/false comparison, we'll use EqualFold.
func CompareStrings(a, b string) int {
	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)

	if aLower == bLower {
		return 0
	}

	if aLower < bLower {
		return -1
	}

	return 1
}

// MatchesFilter tests for match scenarios, where value can contain pattern, value can equal pattern,
// value can start with pattern, or pattern can be a regex expression applied to value
//
// General guidelines...
//
//   - pattern starts with a !!, then the rest of pattern is a regex expression
//   - pattern starts AND ends with an *, then the text between the *'s should be contained in value
//   - pattern only ends in an *, leading portion of value must match the corresponding leading portion of pattern prior to the *
//   - pattern only starts with an *, then value must end with the portion of pattern up to the *, but does not have to exactly MATCH pattern
//   - pattern does NOT start with !! and has no * at end or beginning, then value must case-insensitive MATCH pattern
//
// This does NOT support...
//   - abstract matches (*) in the middle of the match pattern
//   - Placeholder wildcards using ?, but might add that in the future
//
// If the pattern start with a single "!", this is considered a NOT indicator, so that
// match logic is reserved post match of the remainder of the pattern text.
func MatchesFilter(value, pattern string) bool {
	var (
		valueFragment string
		valueLower    = strings.ToLower(value)
	)

	if len(pattern) == 0 {
		return false
	}

	if pattern == "*" || pattern == "**" {
		// ** is prob a typo without the inclusive text, but... we'll treat it as the same as a single *
		return true
	}

	if strings.Index(pattern, "!!") == 0 {
		// regex match
		valueFragment = strings.ToLower(pattern[2:])

		if len(valueFragment) == 0 {
			// did they only provide the "!!"?
			return false
		}

		regexpObj, err := regexp.Compile(valueFragment)
		if err != nil {
			return false
		}

		return regexpObj.MatchString(valueLower)
	}

	useNotLogic := false
	if strings.HasPrefix(pattern, "!") { // not regex, since tested above
		// a NOT condition
		useNotLogic = true

		if len(pattern) == 1 {
			// Someone passed just the "!"?  We'll call this false like "!*"
			return false
		}

		pattern = pattern[1:]
	}

	if len(pattern) > 1 && pattern[0:1] == "*" && pattern[len(pattern)-1:] == "*" {
		// leading AND trailing * == contains
		valueFragment = strings.ToLower(pattern[1 : len(pattern)-1])
		if useNotLogic {
			return !strings.Contains(valueLower, valueFragment)
		}

		return strings.Contains(valueLower, valueFragment)
	}

	// for these next two checks, if there's an * at end or beginning, we know it's longer than 1 char and
	// start and end can't both be *, due to checks above

	if pattern[0:1] == "*" {
		// ends with
		valueFragment = strings.ToLower(pattern[1:])

		if len(valueLower) < len(valueFragment) {
			return false
		}

		if useNotLogic {
			return !strings.HasSuffix(valueLower, valueFragment)
		} else {
			return strings.HasSuffix(valueLower, valueFragment)
		}
	}

	if pattern[len(pattern)-1:] == "*" {
		// starts with
		valueFragment = strings.ToLower(pattern[:len(pattern)-1])

		if len(valueLower) < len(valueFragment) {
			return false
		}

		if useNotLogic {
			return !strings.HasPrefix(valueLower, valueFragment)
		}

		return strings.HasPrefix(valueLower, valueFragment)
	}

	// if none of the above are true, assume we're doing a case insensitive exact match
	if strings.ToLower(value) == strings.ToLower(pattern) {
		if useNotLogic {
			return false
		} else {
			return true
		}
	} else {
		if useNotLogic {
			return true
		} else {
			return false
		}
	}
}
