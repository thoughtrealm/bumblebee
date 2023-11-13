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
	TextWriterTargetBuffered
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
	outputBuffer      *bytes.Buffer
	onStart           TextWriterEventFunc
	afterFlush        TextWriterEventFunc
	startExecuted     bool
	isBuffered        bool
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

	if outputTarget == TextWriterTargetClipboard || outputTarget == TextWriterTargetBuffered {
		newTextWriter.isBuffered = true
	}

	if newTextWriter.isBuffered {
		newTextWriter.outputBuffer = bytes.NewBuffer(nil)
	}

	return newTextWriter
}

// OutputBuffer can only be called once, after which it will return empty bytes.
// This is because of how the bytes buffer works relating to calling "Bytes()" on the buffer.
// This includes when using the clipboard for output.  After flush, the buffer is drained to
// the clipboard.  This would then return nothing.  So, only use this for buffered scenarios that
// are NOT related to the clipboard.  And then ONLY after calling flush.
func (tw *TextWriter) PostFlushOutputBuffer() []byte {
	return tw.outputBuffer.Bytes()
}

func (tw *TextWriter) Write(p []byte) (n int, err error) {
	// Todo: we need some kind of protection logic for text mode outputs, to make sure it is text output safe
	if !tw.startExecuted {
		tw.startExecuted = true
		if tw.onStart != nil {
			tw.onStart()
		}
	}

	if !tw.headerPrinted && tw.headerText != "" {
		headerText := tw.headerText + " " + strings.Repeat("=", tw.calcLineWidth()-(len(tw.headerText)+1))
		err = tw.outputTextLine([]byte(headerText))
		if err != nil {
			return 0, err
		}

		tw.headerPrinted = true
	}

	// First, handle text mode functionality
	if tw.mode == TextWriterModeText {
		inputText := string(p)
		if tw.partialLine != nil {
			inputText = string(tw.partialLine) + inputText
		}

		if !strings.Contains(inputText, "\n") {
			tw.partialLine = []byte(inputText)
			return len(p), nil
		}

		inputLines := strings.Split(inputText, "\n")
		for i := 0; i < len(inputLines); i++ {
			if i == len(inputLines)-1 {
				if !strings.HasSuffix(inputText, "\n") {
					tw.partialLine = []byte(inputLines[i])
					return len(p), nil
				}
			}

			err = tw.outputLine([]byte(inputLines[i]))
			if err != nil {
				return 0, err
			}
		}

		return len(p), nil
	}

	// Now, handle binary mode
	currentPos := 0
	lenP := len(p)
	if len(tw.partialLine) > 0 {
		if len(tw.partialLine)+lenP < tw.lineWidth {
			tw.partialLine = append(tw.partialLine, p...)
			return len(p), nil
		}

		currentPos = tw.lineWidth - len(tw.partialLine)
		currentLine := append(tw.partialLine, p[:currentPos]...)
		err = tw.outputLine(currentLine)
		if err != nil {
			return 0, err
		}

		tw.partialLine = nil
		n += len(currentLine)
	}

	for {
		if lenP-currentPos <= 0 {
			return lenP, nil
		}

		if lenP-currentPos < tw.lineWidth {
			tw.partialLine = p[currentPos:]
			return lenP, nil
		}

		err = tw.outputLine(p[currentPos : currentPos+tw.lineWidth])
		if err != nil {
			return 0, err
		}

		n += tw.lineWidth
		currentPos += tw.lineWidth
	}
}

func (tw *TextWriter) outputLine(line []byte) error {
	var outputText string
	var err error
	switch tw.mode {
	case TextWriterModeBinary:
		outputText = fmt.Sprintf("%02x\n", line)
	case TextWriterModeText:
		outputText = fmt.Sprintf("%s\n", string(line))
	}

	// we need the line feeds for the output buffer, but not the console writes.
	if tw.isBuffered {
		tw.outputBuffer.Write([]byte(outputText))
	} else {
		fmt.Printf(outputText)
	}

	if err == nil {
		tw.totalBytesWritten += len(line)
	}

	return err
}

func (tw *TextWriter) outputTextLine(line []byte) (err error) {
	outputText := fmt.Sprintf("%s\n", string(line))

	if tw.isBuffered {
		tw.outputBuffer.Write([]byte(outputText))
	} else {
		fmt.Println(string(line))
	}

	return err
}

func (tw *TextWriter) calcLineWidth() int {
	if tw.mode == TextWriterModeBinary {
		return tw.lineWidth * 2
	}
	return tw.lineWidth
}

func (tw *TextWriter) PrintFooter() error {
	if !tw.footerPrinted {
		tw.footerPrinted = true
		if tw.footerText != "" {
			footerText := tw.footerText + " " + strings.Repeat("=", tw.calcLineWidth()-(len(tw.footerText)+1))
			err := tw.outputTextLine([]byte(footerText))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tw *TextWriter) PrintTextLine(textLine string) error {
	return tw.outputTextLine([]byte(textLine))
}

func (tw *TextWriter) Flush() (n int, err error) {
	defer func() {
		if tw.afterFlush != nil {
			tw.afterFlush()
		}
	}()

	if len(tw.partialLine) > 0 {
		err = tw.outputLine(tw.partialLine)
		if err != nil {
			return tw.totalBytesWritten, err
		}
	}

	if !tw.footerPrinted {
		err = tw.PrintFooter()
		if tw.outputTarget == TextWriterTargetClipboard {
			err = WriteToClipboard(tw.outputBuffer.Bytes())
			if err != nil {
				return tw.totalBytesWritten, err
			}
		}
	}

	return tw.totalBytesWritten, nil
}

func (tw *TextWriter) Reset(headerText, footerText string) {
	tw.headerPrinted = false
	tw.footerPrinted = false
	tw.headerText = headerText
	tw.footerText = footerText
}
