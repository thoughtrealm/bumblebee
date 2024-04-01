package helpers

import (
	"errors"
	"fmt"
)

func FormatErrorOutputs(err error) string {
	errList := []error{}

	unwrapped := errors.Unwrap(err)
	for {
		if unwrapped == nil {
			break
		}

		errList = append(errList, unwrapped)
		unwrapped = errors.Unwrap(unwrapped)
	}

	if len(errList) == 0 {
		// can this happen?
		return fmt.Sprintf(": %v", err)
	}

	if len(errList) == 1 {
		return fmt.Sprintf(": %v", err)
	}

	var (
		indent  = "  "
		trailer = "..."
		errText = "...\n"
	)

	for i := 0; i < len(errList); i++ {
		if i == len(errList)-1 {
			trailer = ""
		}

		errText += fmt.Sprintf("%s%v%s\n", indent, errList[i], trailer)
		indent += "  "
	}

	return errText
}
