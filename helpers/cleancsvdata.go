package helpers

import (
	"regexp"
	"strings"
)

func cleanCSVData(data string) string {
	// Normalize line endings (convert \r\n and \r to \n)
	data = strings.ReplaceAll(data, "\r\n", "\n")
	data = strings.ReplaceAll(data, "\r", "\n")

	// Split into lines for processing
	lines := strings.Split(data, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove or escape problematic characters
		// Replace unescaped quotes that are not at field boundaries
		// This regex finds quotes that are not properly escaped
		re := regexp.MustCompile(`([^,"])"([^,"])`)
		line = re.ReplaceAllString(line, `$1""$2`)

		// Remove any non-printable characters except tabs
		line = strings.Map(func(r rune) rune {
			if r == '\t' || (r >= 32 && r < 127) || r >= 160 {
				return r
			}
			return -1 // Remove the character
		}, line)

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}
