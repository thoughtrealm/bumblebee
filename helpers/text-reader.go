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
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/thoughtrealm/bumblebee/logger"
	"strconv"
	"strings"
)

var dataBlockMarkers = []string{
	":data",
	":export-user",
	":export-api-key",
}

// TODO: This scanner/parser is NOT the most efficient or highly performing code... maybe do something else in the future

// TextScanner will scan text input and parse bee encrypted data.  It will provide a reader interface for that parsed data.
//   - if no start/end tokens are encountered, it will assume the input is a single, combined data blob and serve accordingly from the read interface.
//
// - This is NOT intended to handle vary large datasets.
type TextScanner struct {
	readBuff *bytes.Buffer
}

// NewTextScanner will return an initialized text scanner.
// Data is optional.  If provided, it will be passed to the Parse method.
func NewTextScanner(data []byte) (*TextScanner, error) {
	logger.Debug("Initializing new text scanner")
	newTextScanner := &TextScanner{readBuff: bytes.NewBuffer(nil)}
	if data != nil {
		logger.Debug("Data provided to NewTextScanner().  Parsing.")
		return newTextScanner, newTextScanner.Parse(data)
	}

	return newTextScanner, nil
}

// Parse will try to determine the nature of the data and parse accordingly.
// It uses this logic:
//   - Does it contain the hex encoding marker ":start"?  If so, parse as hexencoded
//   - Does it break into lines that are only valid hex characters, ignoring marker lines?  If so, parse as one hex encoded combined blob
//   - Otherwise, parse as a binary blob
func (ts *TextScanner) Parse(data []byte) error {
	logger.Debug("Starting text scanner parse")
	logger.Debugfln("length of input data: %d", len(data))
	// logger.Debugfln("%q\n", data)
	if bytes.Contains(data, []byte(":start")) {
		logger.Debug("Parsing as encoded text")
		return ts.tryParseTextEncoding(data)
	}

	if ts.isWindowsDecimalLines(data) {
		logger.Debug("Parsing as Windows Decimal ByteStream from Get-Content AsByteStream/EncodingRaw output")
		return ts.tryParseWindowsDecimalLines(data)
	}

	if ts.isHexBytes(data) {
		logger.Debug("Parsing as hex")
		return ts.tryParseHex(data)
	}

	logger.Debug("Parsing as blob")
	return ts.loadBlob(data)
}

func (ts *TextScanner) isWindowsDecimalLines(data []byte) bool {
	const hexVals = "0123456789"
	lineEnding := ts.detectLineEndingSequence(data)
	lines := bytes.Split(data, lineEnding)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if len(line) > 3 {
			// a single byte won't be great than 3 decimal digits
			return false
		}

		for _, r := range line {
			if !bytes.Contains([]byte(hexVals), []byte{r}) {
				return false
			}
		}
	}

	return true
}

func (ts *TextScanner) isHexBytes(data []byte) bool {
	const hexVals = "ABCDEFabcdef0123456789"
	lineEnding := ts.detectLineEndingSequence(data)
	lines := bytes.Split(data, lineEnding)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if line[0] == ':' {
			continue
		}

		for _, r := range line {
			if !bytes.Contains([]byte(hexVals), []byte{r}) {
				return false
			}
		}
	}

	return true
}

// tryParseHex is only called after we have validated the input to be hex based.
// So we won't check that again.  We just ignore marker lines and grab all hex into one string and decode.
func (ts *TextScanner) tryParseHex(data []byte) error {
	var decodedBytes []byte
	lineEnding := ts.detectLineEndingSequence(data)
	lines := bytes.Split(data, lineEnding)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if line[0] == ':' {
			continue
		}

		dst := make([]byte, hex.DecodedLen(len(line)))
		_, err := hex.Decode(dst, line)
		if err != nil {
			return err
		}

		decodedBytes = append(decodedBytes, dst...)
	}

	ts.readBuff.Write(decodedBytes)

	return nil
}

// tryParseWindowsLine is only called after we have validated the input to conform to a pattern that would be
// returned from a powershell Get-Content call, either from PS <= 5.x using "-Encoding Byte -Raw" or
// new PS project >= 5.x using "-AsByteStream".  This assumes lines of individual decimal values representing
// individual byte values.
func (ts *TextScanner) tryParseWindowsDecimalLines(data []byte) error {
	var decodedBytes []byte
	lineEnding := ts.detectLineEndingSequence(data)
	lines := bytes.Split(data, lineEnding)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		val, err := strconv.Atoi(string(line))
		if err != nil {
			return fmt.Errorf("failed attempting to process input data as Windows decimal lines: %w", err)
		}

		decodedBytes = append(decodedBytes, []byte{byte(val)}...)
	}

	ts.readBuff.Write(decodedBytes)

	return nil
}

func (ts *TextScanner) loadBlob(data []byte) error {
	_, err := ts.readBuff.Write(data)
	return err
}

func (ts *TextScanner) tryParseTextEncoding(data []byte) error {
	type rawEncoding int
	const (
		rawEncodingHex rawEncoding = iota
		rawEncodingBase64
	)

	type parseMode int
	const (
		parseModeUnknown parseMode = iota
		parseModeHeader
		parseModeData
		parseModeCombined
		parseModeRaw
	)

	// we use 3 different []byte vars, because the user MIGHT have rearranged the block sequences.
	// so, we split them out into 3 vars and re-assemble, instead of assuming they are in correct order
	// Todo: This is crude and barbaric... do it differently one day?
	var (
		mode          parseMode
		encoding      rawEncoding
		hdrBytes      []byte
		dataBytes     []byte
		combinedBytes []byte
		rawBytes      []byte
	)

	detectedLineEnding := ts.detectLineEndingSequence(data)
	lines := bytes.Split(data, detectedLineEnding)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// LineStr contains the unchanged input bytes that are useful for correctly converting to bytes for base64 and hex.
		// We trim it and convert to string for easier handling later.
		lineStr := strings.Trim(string(line), " \n\t\r")

		// lineStrLower is a lower cased version of lineStr that optimizes flag value checks without having to
		// lower case lineStr every time.  DO NOT USE lineStrLower for converting values, as base64 will not return
		// correct decoded values after lower casing the inputs.
		lineStrLower := strings.ToLower(lineStr)
		if strings.Contains(lineStrLower, ":start") {
			if strings.Contains(lineStrLower, ":header+data") {
				logger.Debug("text reader: header+data block detected")
				mode = parseModeCombined
				continue
			}

			if strings.Contains(lineStrLower, ":header") {
				logger.Debug("text reader: header block detected")
				mode = parseModeHeader
				continue
			}
			if isMarkerLine, markerText := ts.isDataMarkerLine(lineStrLower); isMarkerLine {
				logger.Debugfln("text reader: data block detected by marker \"%s\"", markerText)
				mode = parseModeData
				continue
			}

			/*
				if strings.Contains(lineStrLower, ":data") {
					logger.Debug("text reader: data block detected")
					mode = parseModeData
					continue
				}

				if strings.Contains(lineStrLower, ":export-user") {
					logger.Debug("text reader: data block detected")
					mode = parseModeData
					continue
				}
			*/

			if strings.Contains(lineStrLower, ":raw") {
				logger.Debug("text reader: raw block detected")
				mode = parseModeRaw

				if strings.Contains(lineStrLower, ":base64") {
					logger.Debug("text reader: raw encoding base64 block detected")
					encoding = rawEncodingBase64
					continue
				}

				if strings.Contains(lineStrLower, ":hex") {
					logger.Debug("text reader: raw encoding hex block detected")
					encoding = rawEncodingHex
					continue
				}

				return errors.New("unknown raw encoding: neither hex nor base64 flags were detected")
			}

			return errors.New("Unknown marker line.  Beings with start marker, but no section indicator found")
		}

		if strings.Contains(lineStrLower, ":end") {
			// we reset mode to unknown at the end of section, in case a start marker is missing for a future section
			mode = parseModeUnknown
			continue
		}

		if mode == parseModeRaw && encoding == rawEncodingBase64 {
			// The reader that emits raw mode output should NEVER have noise lines.
			// So, we don't check for NON base64 lines.  It should always be a base64 line when encoded as base64.
			lineBytes, err := base64.RawStdEncoding.DecodeString(lineStr)
			if err != nil {
				return err
			}

			rawBytes = append(rawBytes, lineBytes...)
			continue
		}

		// After handling raw base64 above, all other formats are hex.
		// So, we check to confirm that it is hex info...  if not we ignore it.
		// We do this because it is possible the user has copied some random noise accidentally or intentionally
		if !ts.isHexBytes(line) {
			continue
		}

		lineBytes, err := hex.DecodeString(lineStr)
		if err != nil {
			return err
		}

		switch mode {
		case parseModeCombined:
			combinedBytes = append(combinedBytes, lineBytes...)
		case parseModeHeader:
			hdrBytes = append(hdrBytes, lineBytes...)
		case parseModeData:
			dataBytes = append(dataBytes, lineBytes...)
		case parseModeRaw:
			rawBytes = append(rawBytes, lineBytes...)
		default:
			return errors.New("unkown parse mode detected during scan of text decoding")
		}
	}

	// We just append all byte vars, even though only one or two will have values.
	// Todo: This is just easier to read, IMO, than a buch of if constructions.  Maybe there is a cleaner way.
	decodedBytes := rawBytes
	decodedBytes = append(decodedBytes, hdrBytes...)
	decodedBytes = append(decodedBytes, dataBytes...)
	decodedBytes = append(decodedBytes, combinedBytes...)

	logger.Debugfln("Len decodedBytes: %d", len(decodedBytes))

	ts.readBuff.Write(decodedBytes)

	return nil
}

func (ts *TextScanner) detectLineEndingSequence(data []byte) []byte {
	windowsLineEnding := []byte("\r\n")
	everyOtherOSLineEnding := []byte("\n")
	if !bytes.Contains(data, everyOtherOSLineEnding) {
		// There are no actual line breaks detect for either Windows or ANYTHING ELSE, so
		// just return the default OS detected line ending sequence.
		return []byte(LineBreak)
	}

	if bytes.Count(data, windowsLineEnding) >= 2 {
		// There should be at least 2 lines for this check.  Hopefully, by checking
		// for two or more counts, we avoid accidental collisions with random stream sequences.
		return windowsLineEnding
	}

	return everyOtherOSLineEnding
}

func (ts *TextScanner) isDataMarkerLine(lineStrLower string) (isMarkerLine bool, markerTextFound string) {
	for _, markerText := range dataBlockMarkers {
		if strings.Contains(lineStrLower, markerText) {
			return true, markerTextFound
		}
	}

	return false, ""
}

func (ts *TextScanner) Read(p []byte) (n int, err error) {
	return ts.readBuff.Read(p)
}
