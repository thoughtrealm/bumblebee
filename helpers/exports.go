package helpers

import "strings"

type ImportInputSource int

const (
	ImportInputSourcePiped ImportInputSource = iota
	ImportInputSourceClipboard
	ImportInputSourceFile
	ImportInputSourceUnknown
)

func TextToImportInputSource(text string) ImportInputSource {
	switch strings.ToUpper(strings.Trim(text, " \t\n\r")) {
	case "PIPED":
		return ImportInputSourcePiped
	case "CLIPBOARD":
		return ImportInputSourceClipboard
	case "FILE":
		return ImportInputSourceFile
	default:
		return ImportInputSourceUnknown
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
