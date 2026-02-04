package helpers

import (
	"regexp"
	"strings"
)

func CleanCSVData(data string) string {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	data = strings.ReplaceAll(data, "\r", "\n")

	lines := strings.Split(data, "\n")
	var cleanedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		re := regexp.MustCompile(`([^,"])"([^,"])`)
		line = re.ReplaceAllString(line, `$1""$2`)

		line = strings.Map(func(r rune) rune {
			if r == '\t' || (r >= 32 && r < 127) || r >= 160 {
				return r
			}
			return -1
		}, line)

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}
