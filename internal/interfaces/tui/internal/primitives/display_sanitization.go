package primitives

import (
	"strings"
	"unicode/utf8"
)

type DisplaySanitizationMode int

const (
	DisplaySanitizeSingleLine DisplaySanitizationMode = iota
	DisplaySanitizeMultiline
)

func SanitizeDisplayText(text string, mode DisplaySanitizationMode) string {
	if text == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(text))

	for i := 0; i < len(text); {
		switch text[i] {
		case '\x1b':
			i = skipTerminalEscapeSequence(text, i)
		case '\r':
			if i+1 < len(text) && text[i+1] == '\n' {
				writeSanitizedLineBreak(&builder, mode)
				i += 2
				continue
			}
			builder.WriteByte(' ')
			i++
		case '\n':
			writeSanitizedLineBreak(&builder, mode)
			i++
		case '\t':
			builder.WriteByte(' ')
			i++
		default:
			if text[i] < 0x20 || text[i] == 0x7f {
				i++
				continue
			}

			_, size := utf8.DecodeRuneInString(text[i:])
			if size < 1 {
				size = 1
			}
			builder.WriteString(text[i : i+size])
			i += size
		}
	}

	return builder.String()
}

func writeSanitizedLineBreak(builder *strings.Builder, mode DisplaySanitizationMode) {
	if mode == DisplaySanitizeMultiline {
		builder.WriteByte('\n')
		return
	}
	builder.WriteByte(' ')
}

func skipTerminalEscapeSequence(text string, start int) int {
	if start+1 >= len(text) {
		return start + 1
	}

	switch text[start+1] {
	case '[':
		return skipCSISequence(text, start+2)
	case ']':
		return skipOSCSequence(text, start+2)
	default:
		return skipESCSequence(text, start+1)
	}
}

func skipCSISequence(text string, index int) int {
	for index < len(text) {
		if text[index] >= 0x40 && text[index] <= 0x7e {
			return index + 1
		}
		index++
	}
	return index
}

func skipOSCSequence(text string, index int) int {
	for index < len(text) {
		switch text[index] {
		case '\a':
			return index + 1
		case '\x1b':
			if index+1 < len(text) && text[index+1] == '\\' {
				return index + 2
			}
		}
		index++
	}
	return index
}

func skipESCSequence(text string, index int) int {
	for index < len(text) && text[index] >= 0x20 && text[index] <= 0x2f {
		index++
	}
	if index < len(text) && text[index] >= 0x30 && text[index] <= 0x7e {
		return index + 1
	}
	return index
}
