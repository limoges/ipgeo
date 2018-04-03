package inmem

import (
	"bufio"
	"io"
	"log"
	"strings"
)

// defaultRecord models a csv record, which is a single line. It provides a single
// Get method which provides an empty string if the record is empty.
type defaultRecord []string

// Get returns the column's value from the record, or an empty string if the
// column has no value.
func (r defaultRecord) Get(i int) string {
	if i >= len(r) {
		return ""
	}
	return r[i]
}

// parseCSV allows simple parsing of a line-based CSV file. It uses a scanner
// as opposed to encoding/csv.Reader because the encoding/csv package has many
// specific behaviours which are not desirable; e.g. multi-line parsing for
// quote columns.
// This CSV parser removes "' characters from lines before parsing.
func parseCSV(r io.Reader, appendLine func([]string) error) error {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	linecleaner := strings.NewReplacer(`"`, ``, `'`, ``)
	var lineno int
	for s.Scan() {
		lineno++
		line := linecleaner.Replace(s.Text())
		record := strings.Split(line, ",")
		if err := appendLine(record); err != nil {
			log.Printf("parseCSV: line %d: %s: %s", lineno, err, line)
			continue
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	return nil
}
