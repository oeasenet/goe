package config

import "strings"

// parseEnvFile parses the KEY=value lines of an env file, in order. Blank
// lines and lines starting with # are skipped, as are lines without "=".
//
// A value in double quotes may span lines and understands the escapes \n, \r,
// \t, \" and \\ (any other backslash is kept); a value in single quotes may
// span lines and is taken literally. Text after the closing quote is ignored.
// An unquoted value is trimmed, and quote characters at its ends are dropped —
// the format before quoted values were understood, so existing files read the
// same. A quote that is never closed falls back to that unquoted reading.
func parseEnvFile(content string) [][2]string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var out [][2]string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if value != "" && (value[0] == '"' || value[0] == '\'') {
			if parsed, extraLines, closed := readQuoted(value, lines[i+1:]); closed {
				out = append(out, [2]string{key, parsed})
				i += extraLines
				continue
			}
		}
		out = append(out, [2]string{key, strings.Trim(value, `"'`)})
	}
	return out
}

// readQuoted reads the quoted value value starts with, continuing onto next
// when the quote is not closed on its own line. It returns the value, how many
// lines of next it used, and whether the quote was closed at all.
func readQuoted(value string, next []string) (string, int, bool) {
	quote := value[0]
	text := value[1:]
	var b strings.Builder
	for used := 0; ; used++ {
		for j := 0; j < len(text); j++ {
			ch := text[j]
			if quote == '"' && ch == '\\' && j+1 < len(text) {
				j++
				switch text[j] {
				case 'n':
					b.WriteByte('\n')
				case 'r':
					b.WriteByte('\r')
				case 't':
					b.WriteByte('\t')
				case '"', '\\':
					b.WriteByte(text[j])
				default:
					b.WriteByte('\\')
					b.WriteByte(text[j])
				}
				continue
			}
			if ch == quote {
				return b.String(), used, true
			}
			b.WriteByte(ch)
		}
		if used >= len(next) {
			return "", 0, false
		}
		b.WriteByte('\n')
		text = next[used]
	}
}
