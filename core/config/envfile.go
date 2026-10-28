package config

import "strings"

// parseEnvFile parses the KEY=value lines of an env file, in order. Blank
// lines and lines starting with # are skipped, as are lines without "=".
//
// A value in double quotes may span lines and understands the escapes \n, \r,
// \t, \" and \\ (any other backslash is kept); a value in single quotes may
// span lines and is taken literally. A comment may follow the closing quote
// after a space. An unquoted value is trimmed, and quote characters at its
// ends are dropped — the format before quoted values were understood. A value
// whose quotes do not enclose it, because the quote is never closed or more
// than a comment follows it, is read that way too, so existing files read the
// same; its key is returned in unbalanced.
func parseEnvFile(content string) (values [][2]string, unbalanced []string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
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
			if parsed, extraLines, ok := readQuoted(value, lines[i+1:]); ok {
				values = append(values, [2]string{key, parsed})
				i += extraLines
				continue
			}
			unbalanced = append(unbalanced, key)
		}
		values = append(values, [2]string{key, strings.Trim(value, `"'`)})
	}
	return values, unbalanced
}

// readQuoted reads the quoted value value starts with, continuing onto next
// when the quote is not closed on its own line. It returns the value, how many
// lines of next it used, and whether the quotes enclose the value: closed, with
// nothing but spaces or a comment after the closing quote.
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
				return b.String(), used, onlyComment(text[j+1:])
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

// onlyComment reports whether rest, the text after a closing quote, is blank
// or a comment: # after at least one space.
func onlyComment(rest string) bool {
	after := strings.TrimLeft(rest, " \t")
	return strings.TrimSpace(after) == "" || (len(after) < len(rest) && after[0] == '#')
}
