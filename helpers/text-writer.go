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
	"fmt"
	"strings"
)

type TextWriterMode int

const (
	TextWriterModeBinary TextWriterMode = iota
	TextWriterModeText
)

type TextWriterTarget int

const (
	TextWriterTargetConsole TextWriterTarget = iota
	TextWriterTargetClipboard
)

type TextWriterEventFunc func()

var NilTextWriterEventFunc TextWriterEventFunc = nil

type TextWriter struct {
	lineWidth         int
	mode              TextWriterMode
	partialLine       []byte
	headerPrinted     bool
	footerPrinted     bool
	totalBytesWritten int
	headerText        string
	footerText        string
	outputTarget      TextWriterTarget
	clipboardBuffer   *bytes.Buffer
	onStart           TextWriterEventFunc
	afterFlush        TextWriterEventFunc
	startExecuted     bool
}

func NewTextWriter(
	outputTarget TextWriterTarget,
	lineWidth int,
	mode TextWriterMode,
	headerText,
	footerText string,
	onStartFunc, afterFlushFunc TextWriterEventFunc,
) *TextWriter {
	newTextWriter := &TextWriter{
		outputTarget: outputTarget,
		lineWidth:    lineWidth,
		mode:         mode,
		headerText:   headerText,
		footerText:   footerText,
		onStart:      onStartFunc,
		afterFlush:   afterFlushFunc,
	}

	if outputTarget == TextWriterTargetClipboard {
		newTextWriter.clipboardBuffer = bytes.NewBuffer(nil)
	}

	return newTextWriter
}

func (bsw *TextWriter) Write(p []byte) (n int, err error) {
	// Todo: we need some kind of protection logic for text mode outputs, to make sure it is text output safe
	if !bsw.startExecuted {
		bsw.startExecuted = true
		if bsw.onStart != nil {
			bsw.onStart()
		}
	}

	if !bsw.headerPrinted && bsw.headerText != "" {
		headerText := bsw.headerText + " " + strings.Repeat("=", bsw.calcLineWidth()-(len(bsw.headerText)+1))
		err = bsw.outputTextLine([]byte(headerText))
		if err != nil {
			return 0, err
		}

		bsw.headerPrinted = true
	}

	// First, handle text mode functionality
	if bsw.mode == TextWriterModeText {
		inputText := string(p)
		if bsw.partialLine != nil {
			inputText = string(bsw.partialLine) + inputText
		}

		if !strings.Contains(inputText, "\n") {
			bsw.partialLine = []byte(inputText)
			return len(p), nil
		}

		inputLines := strings.Split(inputText, "\n")
		for i := 0; i < len(inputLines); i++ {
			if i == len(inputLines)-1 {
				if !strings.HasSuffix(inputText, "\n") {
					bsw.partialLine = []byte(inputLines[i])
					return len(p), nil
				}
			}

			err = bsw.outputLine([]byte(inputLines[i]))
			if err != nil {
				return 0, err
			}
		}

		return len(p), nil
	}

	// Now, handle binary mode
	currentPos := 0
	lenP := len(p)
	if len(bsw.partialLine) > 0 {
		if len(bsw.partialLine)+lenP < bsw.lineWidth {
			bsw.partialLine = append(bsw.partialLine, p...)
			return len(p), nil
		}

		currentPos = bsw.lineWidth - len(bsw.partialLine)
		currentLine := append(bsw.partialLine, p[:currentPos]...)
		err = bsw.outputLine(currentLine)
		if err != nil {
			return 0, err
		}

		bsw.partialLine = nil
		n += len(currentLine)
	}

	for {
		if lenP-currentPos <= 0 {
			return lenP, nil
		}

		if lenP-currentPos < bsw.lineWidth {
			bsw.partialLine = p[currentPos:]
			return lenP, nil
		}

		err = bsw.outputLine(p[currentPos : currentPos+bsw.lineWidth])
		if err != nil {
			return 0, err
		}

		n += bsw.lineWidth
		currentPos += bsw.lineWidth
	}
}

func (bsw *TextWriter) outputLine(line []byte) error {
	var outputText string
	var err error
	switch bsw.mode {
	case TextWriterModeBinary:
		outputText = fmt.Sprintf("%02x\n", line)
	case TextWriterModeText:
		outputText = fmt.Sprintf("%s\n", string(line))
	}

	// we need the line feeds for the clipboard buffer, but not the console writes.
	switch bsw.outputTarget {
	case TextWriterTargetConsole:
		fmt.Printf(outputText)
	case TextWriterTargetClipboard:
		bsw.clipboardBuffer.Write([]byte(outputText))
	}

	if err == nil {
		bsw.totalBytesWritten += len(line)
	}

	return err
}

func (bsw *TextWriter) outputTextLine(line []byte) (err error) {
	outputText := fmt.Sprintf("%s\n", string(line))

	switch bsw.outputTarget {
	case TextWriterTargetConsole:
		fmt.Println(string(line))
	case TextWriterTargetClipboard:
		bsw.clipboardBuffer.Write([]byte(outputText))
	}

	return err
}

func (bsw *TextWriter) calcLineWidth() int {
	if bsw.mode == TextWriterModeBinary {
		return bsw.lineWidth * 2
	}
	return bsw.lineWidth
}

func (bsw *TextWriter) PrintFooter() error {
	if !bsw.footerPrinted {
		bsw.footerPrinted = true
		if bsw.footerText != "" {
			footerText := bsw.footerText + " " + strings.Repeat("=", bsw.calcLineWidth()-(len(bsw.footerText)+1))
			err := bsw.outputTextLine([]byte(footerText))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bsw *TextWriter) PrintTextLine(textLine string) error {
	return bsw.outputTextLine([]byte(textLine))
}

func (bsw *TextWriter) Flush() (n int, err error) {
	defer func() {
		if bsw.afterFlush != nil {
			bsw.afterFlush()
		}
	}()

	if len(bsw.partialLine) > 0 {
		err = bsw.outputLine(bsw.partialLine)
		if err != nil {
			return bsw.totalBytesWritten, err
		}
	}

	if !bsw.footerPrinted {
		err = bsw.PrintFooter()
		if bsw.outputTarget == TextWriterTargetClipboard {
			err = WriteToClipboard(bsw.clipboardBuffer.Bytes())
			if err != nil {
				return bsw.totalBytesWritten, err
			}
		}
	}

	return bsw.totalBytesWritten, nil
}

func (bsw *TextWriter) Reset(headerText, footerText string) {
	bsw.headerPrinted = false
	bsw.footerPrinted = false
	bsw.headerText = headerText
	bsw.footerText = footerText
}
