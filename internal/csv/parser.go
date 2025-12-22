// Package csv provides utilities for reading and writing CSV files.
package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// MaxFileSize is the maximum allowed CSV file size (100MB).
// Prevents OOM from maliciously large or accidental huge files.
var MaxFileSize int64 = 100 * 1024 * 1024

// MaxHeaderSearchRows is the maximum number of rows to scan when looking for the CSV header.
// Some CSV exports have metadata rows before the actual header.
var MaxHeaderSearchRows = 20

// sanitizeUTF8 replaces invalid UTF-8 byte sequences with the Unicode replacement character.
// This handles CSV files exported from Excel with Windows-1252 or other legacy encodings.
func sanitizeUTF8(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}

	// Replace invalid sequences with replacement character
	var buf bytes.Buffer
	buf.Grow(len(data))

	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			// Invalid byte - replace with replacement character
			buf.WriteRune('\uFFFD')
			data = data[1:]
		} else {
			buf.WriteRune(r)
			data = data[size:]
		}
	}

	return buf.Bytes()
}

// Read reads all records from a CSV file.
// It checks file size before reading to prevent OOM attacks.
// Invalid UTF-8 byte sequences are replaced with the Unicode replacement character.
func Read(path string) ([][]string, error) {
	// Check file size before reading to prevent OOM
	info, err := os.Stat(path)
	if err != nil {
		return [][]string{}, fmt.Errorf("stat file %q: %w", filepath.Base(path), err)
	}
	if info.Size() > MaxFileSize {
		return [][]string{}, fmt.Errorf("file %q exceeds maximum size (%d MB limit, file is %d MB)",
			filepath.Base(path), MaxFileSize/(1024*1024), info.Size()/(1024*1024))
	}

	// Read entire file and sanitize UTF-8
	data, err := os.ReadFile(path)
	if err != nil {
		return [][]string{}, fmt.Errorf("read file %q: %w", filepath.Base(path), err)
	}
	data = sanitizeUTF8(data)

	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // allow variable row lengths
	r.LazyQuotes = true    // allow bare quotes in fields (common in real-world CSVs)

	records, err := r.ReadAll()
	if err != nil {
		return [][]string{}, fmt.Errorf("parse file %q: %w", filepath.Base(path), err)
	}

	return records, nil
}

// Write writes records to a CSV file.
func Write(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file %q: %w", filepath.Base(path), err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write to csv %q: %w", filepath.Base(path), err)
		}
	}

	// Flush buffer data to disk
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("flush csv %q: %w", filepath.Base(path), err)
	}

	return nil
}

// FindHeaderRow searches for a header row matching the required columns.
// Returns the 0-based row index where the header was found.
// Invalid UTF-8 byte sequences are replaced with the Unicode replacement character.
func FindHeaderRow(path string, required []string) (int, error) {
	// Read entire file and sanitize UTF-8
	data, err := os.ReadFile(path)
	if err != nil {
		return -1, fmt.Errorf("read file %q: %w", filepath.Base(path), err)
	}
	data = sanitizeUTF8(data)

	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // allow variable row lengths
	r.LazyQuotes = true    // allow bare quotes in fields (common in real-world CSVs)

	// Scan up to MaxHeaderSearchRows looking for the header
	for i := 0; i < MaxHeaderSearchRows; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return -1, fmt.Errorf("read row %d: %w", i, err)
		}

		// Exact header match
		if EqualHeaders(record, required) {
			return i, nil
		}
	}

	return -1, fmt.Errorf("header not found within first %d rows", MaxHeaderSearchRows)
}

// EqualHeaders compares two header rows for equality (case-insensitive, cleaned).
func EqualHeaders(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		aa := CleanCell(a[i])
		bb := CleanCell(b[i])

		if !strings.EqualFold(aa, bb) {
			return false
		}
	}
	return true
}

// CleanHeader cleans a header value for use as a map key.
// Applies CleanCell and lowercases the result for case-insensitive matching.
func CleanHeader(s string) string {
	return strings.ToLower(CleanCell(s))
}

// CleanCell removes common CSV artifacts from a cell value:
// - Trims whitespace
// - Removes Excel formula prefix (="...")
// - Removes surrounding quotes
// - Removes "netsuite:" prefix
func CleanCell(s string) string {
	s = strings.TrimSpace(s)

	// Remove leading '='
	if strings.HasPrefix(s, "=\"") && strings.HasSuffix(s, "\"") {
		s = s[2 : len(s)-1] // remove =" and final "
	} else if strings.HasPrefix(s, "=") {
		s = s[1:]
	}

	// Remove any surrounding quotes
	s = strings.Trim(s, `"'`)

	// Remove "netsuite:" prefix if present
	s = strings.TrimPrefix(s, "netsuite:")

	return s
}
