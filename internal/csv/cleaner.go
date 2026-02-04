package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	minPrintableASCII = 32
	maxPrintableASCII = 126
	minExtendedASCII  = 160
)

// Cleaner handles CSV data cleaning operations.
type Cleaner struct {
	unescapedQuoteRegex *regexp.Regexp
}

// NewCleaner creates a new CSV cleaner instance.
func NewCleaner() *Cleaner {
	return &Cleaner{
		unescapedQuoteRegex: regexp.MustCompile(`([^,"])"([^,"])`),
	}
}

// CleanData cleans raw CSV string data by normalizing line endings,
// removing empty lines, escaping unescaped quotes, and filtering
// non-printable characters.
func (c *Cleaner) CleanData(data string) string {
	data = c.normalizeLineEndings(data)
	lines := strings.Split(data, "\n")
	
	cleanedLines := make([]string, 0, len(lines))
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		line = c.escapeUnescapedQuotes(line)
		line = c.filterNonPrintableChars(line)
		
		cleanedLines = append(cleanedLines, line)
	}
	
	return strings.Join(cleanedLines, "\n")
}

// CleanRecords processes CSV records by trimming whitespace,
// removing unwanted characters, filtering empty rows, and
// removing duplicate rows.
func (c *Cleaner) CleanRecords(reader *csv.Reader, writer *csv.Writer) error {
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	
	cleanedHeader := c.cleanRow(header)
	if err := writer.Write(cleanedHeader); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	
	seenRows := make(map[string]struct{})
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read record: %w", err)
		}
		
		cleanedRow := c.cleanRow(record)
		
		if c.isEmptyRow(cleanedRow) {
			continue
		}
		
		rowKey := strings.Join(cleanedRow, "|")
		if _, exists := seenRows[rowKey]; exists {
			continue
		}
		
		seenRows[rowKey] = struct{}{}
		
		if err := writer.Write(cleanedRow); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}
	
	return nil
}

func (c *Cleaner) normalizeLineEndings(data string) string {
	data = strings.ReplaceAll(data, "\r\n", "\n")
	data = strings.ReplaceAll(data, "\r", "\n")
	return data
}

func (c *Cleaner) escapeUnescapedQuotes(line string) string {
	return c.unescapedQuoteRegex.ReplaceAllString(line, `$1""$2`)
}

func (c *Cleaner) filterNonPrintableChars(line string) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' || 
		   (r >= minPrintableASCII && r < maxPrintableASCII) || 
		   r >= minExtendedASCII {
			return r
		}
		return -1
	}, line)
}

func (c *Cleaner) cleanRow(row []string) []string {
	cleaned := make([]string, len(row))
	
	for i, field := range row {
		field = strings.TrimSpace(field)
		field = c.removeNonPrintableChars(field)
		cleaned[i] = field
	}
	
	return cleaned
}

func (c *Cleaner) removeNonPrintableChars(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	
	for _, r := range s {
		if (r >= minPrintableASCII && r <= maxPrintableASCII) || 
		   r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

func (c *Cleaner) isEmptyRow(row []string) bool {
	for _, field := range row {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
