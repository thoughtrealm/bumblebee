package helpers

import "strings"

type ExportInputSource int

const (
	ExportInputSourcePiped ExportInputSource = iota
	ExportInputSourceClipboard
	ExportInputSourceFile
	ExportInputSourceUnknown
)

func TextToExportInputSource(text string) ExportInputSource {
	switch strings.ToUpper(strings.Trim(text, " \t\n\r")) {
	case "PIPED":
		return ExportInputSourcePiped
	case "CLIPBOARD":
		return ExportInputSourceClipboard
	case "FILE":
		return ExportInputSourceFile
	default:
		return ExportInputSourceUnknown
	}
}

type ExportOutputTarget int

const (
	ExportOutputTargetConsole ExportOutputTarget = iota
	ExportOutputTargetClipboard
	ExportOutputTargetFile
	ExportOutputTargetUnknown
)

func TextToExportOutputTarget(text string) ExportOutputTarget {
	switch strings.ToUpper(strings.Trim(text, " \t\n\r")) {
	case "CONSOLE":
		return ExportOutputTargetConsole
	case "CLIPBOARD":
		return ExportOutputTargetClipboard
	case "FILE":
		return ExportOutputTargetFile
	default:
		return ExportOutputTargetUnknown
	}
}

type ExportOutputEncoding int

const (
	ExportOutputEncodingRaw ExportOutputEncoding = iota
	ExportOutputEncodingText
	ExportOutputEncodingUnknown
)

func TextToExportOutputEncoding(text string) ExportOutputEncoding {
	switch strings.ToUpper(strings.Trim(text, " \t\n\r")) {
	case "RAW":
		return ExportOutputEncodingRaw
	case "TEXT":
		return ExportOutputEncodingText
	default:
		return ExportOutputEncodingUnknown
	}
}
