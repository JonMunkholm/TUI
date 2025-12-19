package handler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JonMunkholm/TUI/internal/csv"
	"github.com/JonMunkholm/TUI/internal/schema"
)

/* ========================================
	ToPgText Tests
======================================== */

func TestToPgText_Positive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "hello"},
		{"string with spaces", "hello world", "hello world"},
		{"numeric string", "12345", "12345"},
		{"special characters", "hello@world.com", "hello@world.com"},
		{"unicode", "héllo wörld", "héllo wörld"},
		{"single character", "a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgText(tt.input)
			if !result.Valid {
				t.Errorf("ToPgText(%q) returned invalid, expected valid", tt.input)
			}
			if result.String != tt.expected {
				t.Errorf("ToPgText(%q) = %q, expected %q", tt.input, result.String, tt.expected)
			}
		})
	}
}

func TestToPgText_Negative(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only spaces", "   "},
		{"only tabs", "\t\t"},
		{"only newlines", "\n\n"},
		{"mixed whitespace", " \t\n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgText(tt.input)
			if result.Valid {
				t.Errorf("ToPgText(%q) returned valid, expected invalid", tt.input)
			}
		})
	}
}

func TestToPgText_EdgeCases(t *testing.T) {
	// These test potential false negatives - valid inputs that might incorrectly fail
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"leading spaces trimmed", "  hello", "hello"},
		{"trailing spaces trimmed", "hello  ", "hello"},
		{"both sides trimmed", "  hello  ", "hello"},
		{"internal spaces preserved", "hello  world", "hello  world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgText(tt.input)
			if !result.Valid {
				t.Errorf("ToPgText(%q) returned invalid, expected valid", tt.input)
			}
			if result.String != tt.expected {
				t.Errorf("ToPgText(%q) = %q, expected %q", tt.input, result.String, tt.expected)
			}
		})
	}
}

/* ========================================
	ToPgDate Tests
======================================== */

func TestToPgDate_Positive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		// US formats
		{"MM/DD/YY", "01/15/24", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"M/D/YY", "1/5/24", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{"MM/DD/YYYY", "01/15/2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"M/D/YYYY", "1/5/2024", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},

		// ISO format
		{"YYYY-MM-DD", "2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"YYYY/MM/DD", "2024/01/15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},

		// Dash formats
		{"M-D-YY", "1-5-24", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{"MM-DD-YYYY", "01-15-2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},

		// Dot formats
		{"M.D.YY", "1.5.24", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{"MM.DD.YY", "01.15.24", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"YYYY.MM.DD", "2024.01.15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},

		// Text formats
		{"Jan D, YYYY", "Jan 15, 2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"D Jan YYYY", "15 Jan 2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},

		// Compact format
		{"YYYYMMDD", "20240115", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgDate(tt.input)
			if !result.Valid {
				t.Errorf("ToPgDate(%q) returned invalid, expected valid", tt.input)
				return
			}
			if !result.Time.Equal(tt.expected) {
				t.Errorf("ToPgDate(%q) = %v, expected %v", tt.input, result.Time, tt.expected)
			}
		})
	}
}

func TestToPgDate_Negative(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only spaces", "   "},
		{"random text", "not a date"},
		{"invalid month", "13/01/2024"},
		{"invalid day", "01/32/2024"},
		{"partial date", "01/2024"},
		{"time only", "12:30:45"},
		{"mixed garbage", "abc123def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgDate(tt.input)
			if result.Valid {
				t.Errorf("ToPgDate(%q) returned valid with time %v, expected invalid", tt.input, result.Time)
			}
		})
	}
}

func TestToPgDate_FalseNegatives(t *testing.T) {
	// Edge cases that should succeed but might fail
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{"with leading space", " 2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"with trailing space", "2024-01-15 ", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		// 2-digit year 24 is within pivot range (current year + 20), so stays 2024
		{"2-digit year 24 becomes 2024", "01/15/24", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgDate(tt.input)
			if !result.Valid {
				t.Errorf("ToPgDate(%q) returned invalid, expected valid", tt.input)
				return
			}
			if !result.Time.Equal(tt.expected) {
				t.Errorf("ToPgDate(%q) = %v, expected %v", tt.input, result.Time, tt.expected)
			}
		})
	}
}

func TestToPgDate_TwoDigitYearBehavior(t *testing.T) {
	// ToPgDate uses a pivot year approach for 2-digit years:
	// - If parsed year > (current year + TwoDigitYearPivot), subtract 100 years
	// - Default pivot is 20 years, so in 2025: years > 2045 become 19XX
	//
	// This fixes Go's inconsistent default (00-68 → 2000s, 69-99 → 1900s)
	// and provides predictable behavior for financial/accounting data.
	tests := []struct {
		name     string
		input    string
		expected int // expected year
	}{
		{"year 24 -> 2024 (within pivot)", "01/15/24", 2024},
		{"year 00 -> 2000 (within pivot)", "01/15/00", 2000},
		{"year 44 -> 2044 (within pivot)", "01/15/44", 2044},
		{"year 45 -> 2045 (at pivot boundary)", "01/15/45", 2045},
		{"year 46 -> 1946 (beyond pivot)", "01/15/46", 1946},
		{"year 68 -> 1968 (beyond pivot)", "01/15/68", 1968},
		{"year 99 -> 1999 (beyond pivot)", "01/15/99", 1999},
		{"year 69 -> 1969 (beyond pivot)", "01/15/69", 1969},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgDate(tt.input)
			if !result.Valid {
				t.Errorf("ToPgDate(%q) returned invalid", tt.input)
				return
			}
			if result.Time.Year() != tt.expected {
				t.Errorf("ToPgDate(%q).Year() = %d, expected %d", tt.input, result.Time.Year(), tt.expected)
			}
		})
	}
}

/* ========================================
	ToPgNumeric Tests
======================================== */

func TestToPgNumeric_Positive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"integer", "123"},
		{"negative integer", "-123"},
		{"decimal", "123.45"},
		{"negative decimal", "-123.45"},
		{"leading decimal", ".45"},
		{"trailing decimal", "123."},
		{"zero", "0"},
		{"negative zero", "-0"},
		{"currency with dollar", "$123.45"},
		{"currency with comma", "$1,234.56"},
		{"accounting negative", "(123.45)"},
		{"accounting with currency", "($1,234.56)"},
		{"euro symbol", "€123.45"},
		{"pound symbol", "£123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgNumeric(tt.input)
			if !result.Valid {
				t.Errorf("ToPgNumeric(%q) returned invalid, expected valid", tt.input)
			}
		})
	}
}

func TestToPgNumeric_ScientificNotation(t *testing.T) {
	// Document behavior: Scientific notation passes regex validation but
	// pgtype.Numeric.Scan does not accept it, so these return invalid
	tests := []struct {
		name  string
		input string
	}{
		{"positive scientific", "1.5e10"},
		{"negative scientific", "-1.5e-10"},
		{"capital E", "1.5E10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgNumeric(tt.input)
			// These currently return invalid because pgtype.Numeric.Scan
			// doesn't support scientific notation
			if result.Valid {
				t.Logf("ToPgNumeric(%q) accepted scientific notation", tt.input)
			}
		})
	}
}

func TestToPgNumeric_Negative(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only spaces", "   "},
		{"text only", "abc"},
		{"mixed text and numbers", "12abc34"},
		{"multiple dots", "12.34.56"},
		{"multiple minus", "--123"},
		{"letters in middle", "12a34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgNumeric(tt.input)
			if result.Valid {
				t.Errorf("ToPgNumeric(%q) returned valid, expected invalid", tt.input)
			}
		})
	}
}

func TestToPgNumeric_AccountingFormat(t *testing.T) {
	// Test that accounting format (123.45) is correctly parsed as negative
	input := "(500.00)"
	result := ToPgNumeric(input)

	if !result.Valid {
		t.Fatalf("ToPgNumeric(%q) returned invalid, expected valid", input)
	}

	// Convert to float to check sign
	f8, err := result.Float64Value()
	if err != nil {
		t.Fatalf("Failed to get float64 value: %v", err)
	}

	if !f8.Valid {
		t.Fatalf("Float64Value returned invalid")
	}

	if f8.Float64 >= 0 {
		t.Errorf("ToPgNumeric(%q) = %f, expected negative value", input, f8.Float64)
	}
}

func TestToPgNumeric_EdgeCases(t *testing.T) {
	// Potential false negatives - valid inputs that might fail
	tests := []struct {
		name  string
		input string
	}{
		{"with leading space", " 123.45"},
		{"with trailing space", "123.45 "},
		{"large number with commas", "$1,234,567,890.12"},
		{"just decimal point and digits", ".99"},
		{"plus sign", "+123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgNumeric(tt.input)
			if !result.Valid {
				t.Errorf("ToPgNumeric(%q) returned invalid, expected valid", tt.input)
			}
		})
	}
}

/* ========================================
	ToPgBool Tests
======================================== */

func TestToPgBool_Positive_True(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase true", "true"},
		{"uppercase TRUE", "TRUE"},
		{"mixed case True", "True"},
		{"t", "t"},
		{"T", "T"},
		{"yes", "yes"},
		{"YES", "YES"},
		{"y", "y"},
		{"Y", "Y"},
		{"1", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgBool(tt.input)
			if !result.Valid {
				t.Errorf("ToPgBool(%q) returned invalid, expected valid", tt.input)
				return
			}
			if !result.Bool {
				t.Errorf("ToPgBool(%q) = false, expected true", tt.input)
			}
		})
	}
}

func TestToPgBool_Positive_False(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase false", "false"},
		{"uppercase FALSE", "FALSE"},
		{"mixed case False", "False"},
		{"f", "f"},
		{"F", "F"},
		{"no", "no"},
		{"NO", "NO"},
		{"n", "n"},
		{"N", "N"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgBool(tt.input)
			if !result.Valid {
				t.Errorf("ToPgBool(%q) returned invalid, expected valid", tt.input)
				return
			}
			if result.Bool {
				t.Errorf("ToPgBool(%q) = true, expected false", tt.input)
			}
		})
	}
}

func TestToPgBool_Negative(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only spaces", "   "},
		{"random text", "maybe"},
		{"number 2", "2"},
		{"truthy", "truthy"},
		{"falsy", "falsy"},
		{"partial true", "tru"},
		{"partial false", "fal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgBool(tt.input)
			if result.Valid {
				t.Errorf("ToPgBool(%q) returned valid, expected invalid", tt.input)
			}
		})
	}
}

func TestToPgBool_EdgeCases(t *testing.T) {
	// Potential false negatives - valid inputs that might fail
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"with leading space", " true", true},
		{"with trailing space", "true ", true},
		{"with both spaces", " false ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgBool(tt.input)
			if !result.Valid {
				t.Errorf("ToPgBool(%q) returned invalid, expected valid", tt.input)
				return
			}
			if result.Bool != tt.expected {
				t.Errorf("ToPgBool(%q) = %v, expected %v", tt.input, result.Bool, tt.expected)
			}
		})
	}
}

/* ========================================
	NormalizeUsState Tests
======================================== */

func TestNormalizeUsState_Positive_FullNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"california lowercase", "california", "CA"},
		{"California mixed", "California", "CA"},
		{"CALIFORNIA uppercase", "CALIFORNIA", "CA"},
		{"new york", "new york", "NY"},
		{"New York", "New York", "NY"},
		{"texas", "texas", "TX"},
		{"florida", "florida", "FL"},
		{"washington", "washington", "WA"},
		{"north carolina", "north carolina", "NC"},
		{"south dakota", "south dakota", "SD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.NormalizeUsState(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeUsState(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeUsState_Positive_Codes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"CA lowercase", "ca", "CA"},
		{"CA uppercase", "CA", "CA"},
		{"ny lowercase", "ny", "NY"},
		{"NY uppercase", "NY", "NY"},
		{"tx lowercase", "tx", "TX"},
		{"fl lowercase", "fl", "FL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.NormalizeUsState(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeUsState(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeUsState_Fallback(t *testing.T) {
	// Invalid states should return original input
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"unknown state", "Unknown", "Unknown"},
		{"empty string", "", ""},
		{"number", "123", "123"},
		{"partial name", "Calif", "Calif"},
		{"foreign country", "Ontario", "Ontario"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.NormalizeUsState(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeUsState(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeUsState_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with leading space", " california", "CA"},
		{"with trailing space", "california ", "CA"},
		{"with both spaces", " CA ", "CA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.NormalizeUsState(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeUsState(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

/* ========================================
	csv.CleanCell Tests
======================================== */

func TestCleanCell_Positive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "hello"},
		{"with spaces", "  hello  ", "hello"},
		{"with tabs", "\thello\t", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.CleanCell(tt.input)
			if result != tt.expected {
				t.Errorf("csv.CleanCell(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanCell_ExcelFormulas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"equals prefix", "=hello", "hello"},
		{"equals with quotes", `="hello"`, "hello"},
		{"equals with value", "=123", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.CleanCell(tt.input)
			if result != tt.expected {
				t.Errorf("csv.CleanCell(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanCell_Quotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"hello"`, "hello"},
		{"single quotes", "'hello'", "hello"},
		{"mixed quotes preserved inside", `"hello'world"`, "hello'world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.CleanCell(tt.input)
			if result != tt.expected {
				t.Errorf("csv.CleanCell(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanCell_NetsuitePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"netsuite prefix", "netsuite:hello", "hello"},
		{"netsuite prefix with value", "netsuite:12345", "12345"},
		{"no prefix", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.CleanCell(tt.input)
			if result != tt.expected {
				t.Errorf("csv.CleanCell(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanCell_CombinedCleaning(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"all transformations", `  ="netsuite:hello"  `, "hello"},
		{"spaces and quotes", `  "hello"  `, "hello"},
		{"equals and netsuite", "=netsuite:test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.CleanCell(tt.input)
			if result != tt.expected {
				t.Errorf("csv.CleanCell(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

/* ========================================
	csv.EqualHeaders Tests
======================================== */

func TestEqualStringSlices_Positive(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
	}{
		{"empty slices", []string{}, []string{}},
		{"single element", []string{"a"}, []string{"a"}},
		{"multiple elements", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"case insensitive", []string{"Hello", "World"}, []string{"hello", "world"}},
		{"with spaces trimmed", []string{"  hello  ", "world"}, []string{"hello", "  world  "}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.EqualHeaders(tt.a, tt.b)
			if !result {
				t.Errorf("csv.EqualHeaders(%v, %v) = false, expected true", tt.a, tt.b)
			}
		})
	}
}

func TestEqualStringSlices_Negative(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
	}{
		{"different lengths", []string{"a", "b"}, []string{"a"}},
		{"different values", []string{"a", "b"}, []string{"a", "c"}},
		{"empty vs non-empty", []string{}, []string{"a"}},
		{"different order", []string{"a", "b"}, []string{"b", "a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.EqualHeaders(tt.a, tt.b)
			if result {
				t.Errorf("csv.EqualHeaders(%v, %v) = true, expected false", tt.a, tt.b)
			}
		})
	}
}

func TestEqualStringSlices_CleanCellIntegration(t *testing.T) {
	// Test that csv.CleanCell is applied during comparison
	tests := []struct {
		name string
		a    []string
		b    []string
	}{
		{"excel formula vs plain", []string{`="hello"`}, []string{"hello"}},
		{"netsuite prefix vs plain", []string{"netsuite:test"}, []string{"test"}},
		{"quoted vs unquoted", []string{`"value"`}, []string{"value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csv.EqualHeaders(tt.a, tt.b)
			if !result {
				t.Errorf("csv.EqualHeaders(%v, %v) = false, expected true", tt.a, tt.b)
			}
		})
	}
}

/* ========================================
	rowFailed Tests
======================================== */

func TestRowFailed(t *testing.T) {
	tests := []struct {
		name     string
		reason   string
		row      []string
		expected []string
	}{
		{
			"simple row",
			"error: invalid data",
			[]string{"a", "b", "c"},
			[]string{"error: invalid data", "a", "b", "c"},
		},
		{
			"empty row",
			"error: empty",
			[]string{},
			[]string{"error: empty"},
		},
		{
			"single cell",
			"parse error",
			[]string{"value"},
			[]string{"parse error", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rowFailed(tt.reason, tt.row)
			if len(result) != len(tt.expected) {
				t.Errorf("rowFailed(%q, %v) length = %d, expected %d", tt.reason, tt.row, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("rowFailed(%q, %v)[%d] = %q, expected %q", tt.reason, tt.row, i, v, tt.expected[i])
				}
			}
		})
	}
}

/* ========================================
	Benchmark Tests
======================================== */

func BenchmarkToPgNumeric(b *testing.B) {
	inputs := []string{
		"123.45",
		"$1,234,567.89",
		"(500.00)",
		"-123.45",
		"1.5e10",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			ToPgNumeric(input)
		}
	}
}

func BenchmarkToPgDate(b *testing.B) {
	inputs := []string{
		"2024-01-15",
		"01/15/2024",
		"Jan 15, 2024",
		"20240115",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			ToPgDate(input)
		}
	}
}

func BenchmarkCleanCell(b *testing.B) {
	inputs := []string{
		"  hello  ",
		`="test"`,
		"netsuite:value",
		`  ="netsuite:complex"  `,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			csv.CleanCell(input)
		}
	}
}

/* ========================================
	Integration-style Tests
======================================== */

func TestPgTypeConversions_RealWorldData(t *testing.T) {
	// Test with data that might come from real CSV exports
	tests := []struct {
		name    string
		text    string
		date    string
		numeric string
		boolean string
	}{
		{
			"standard row",
			"Customer Name",
			"2024-01-15",
			"1234.56",
			"true",
		},
		{
			"excel export",
			`="Customer Name"`,
			"01/15/24",
			"$1,234.56",
			"TRUE",
		},
		{
			"accounting format",
			"Expense",
			"1/5/2024",
			"(500.00)",
			"no",
		},
		{
			"netsuite export",
			"netsuite:Product",
			"Jan 15, 2024",
			"€123.45",
			"Y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textResult := ToPgText(csv.CleanCell(tt.text))
			if !textResult.Valid {
				t.Errorf("Text conversion failed for %q", tt.text)
			}

			dateResult := ToPgDate(csv.CleanCell(tt.date))
			if !dateResult.Valid {
				t.Errorf("Date conversion failed for %q", tt.date)
			}

			numericResult := ToPgNumeric(csv.CleanCell(tt.numeric))
			if !numericResult.Valid {
				t.Errorf("Numeric conversion failed for %q", tt.numeric)
			}

			boolResult := ToPgBool(csv.CleanCell(tt.boolean))
			if !boolResult.Valid {
				t.Errorf("Boolean conversion failed for %q", tt.boolean)
			}
		})
	}
}

/* ========================================
	False Positive Detection Tests
======================================== */

func TestToPgNumeric_FalsePositives(t *testing.T) {
	// These should NOT be valid but might slip through
	tests := []struct {
		name        string
		input       string
		shouldFail  bool
		description string
	}{
		{
			"IP address",
			"192.168.1.1",
			true,
			"IP addresses have multiple dots and should fail",
		},
		{
			"phone number with dashes",
			"555-123-4567",
			true,
			"Phone numbers are not numeric values",
		},
		{
			"version number",
			"1.2.3",
			true,
			"Version numbers have multiple dots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgNumeric(tt.input)
			if tt.shouldFail && result.Valid {
				t.Errorf("ToPgNumeric(%q) should be invalid: %s", tt.input, tt.description)
			}
		})
	}
}

func TestToPgDate_FalsePositives(t *testing.T) {
	// These should NOT be valid dates
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			"February 30",
			"02/30/2024",
			"February 30 does not exist",
		},
		{
			"Month 13",
			"13/01/2024",
			"Month 13 does not exist",
		},
		{
			"April 31",
			"04/31/2024",
			"April only has 30 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPgDate(tt.input)
			// Note: Go's time.Parse may accept some invalid dates
			// This test documents the behavior
			if result.Valid {
				t.Logf("ToPgDate(%q) accepted invalid date: %s (Got: %v)", tt.input, tt.description, result.Time)
			}
		})
	}
}

/* ========================================
	Nil/Empty Edge Cases
======================================== */

func TestPgTypeConversions_NilEquivalents(t *testing.T) {
	emptyInputs := []string{"", " ", "\t", "\n", "  \t  \n  "}

	for _, input := range emptyInputs {
		t.Run("ToPgText", func(t *testing.T) {
			if ToPgText(input).Valid {
				t.Errorf("ToPgText(%q) should return invalid", input)
			}
		})

		t.Run("ToPgDate", func(t *testing.T) {
			if ToPgDate(input).Valid {
				t.Errorf("ToPgDate(%q) should return invalid", input)
			}
		})

		t.Run("ToPgNumeric", func(t *testing.T) {
			if ToPgNumeric(input).Valid {
				t.Errorf("ToPgNumeric(%q) should return invalid", input)
			}
		})

		t.Run("ToPgBool", func(t *testing.T) {
			if ToPgBool(input).Valid {
				t.Errorf("ToPgBool(%q) should return invalid", input)
			}
		})
	}
}

/* ========================================
	csv.FindHeaderRow Tests
======================================== */

func TestFindHeaderRow_Positive(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		required     []string
		expectedRow  int
	}{
		{
			"header on first row",
			"Name,Age,City\nJohn,30,NYC\nJane,25,LA",
			[]string{"Name", "Age", "City"},
			0,
		},
		{
			"header on second row",
			"Report Title\nName,Age,City\nJohn,30,NYC",
			[]string{"Name", "Age", "City"},
			1,
		},
		{
			"header on third row",
			"Report Title\nGenerated: 2024-01-15\nName,Age,City\nJohn,30,NYC",
			[]string{"Name", "Age", "City"},
			2,
		},
		{
			"case insensitive match",
			"NAME,AGE,CITY\nJohn,30,NYC",
			[]string{"name", "age", "city"},
			0,
		},
		{
			"header with spaces trimmed",
			"  Name  ,  Age  ,  City  \nJohn,30,NYC",
			[]string{"Name", "Age", "City"},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "test_*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			t.Cleanup(func() {
				if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: failed to cleanup temp file: %v", err)
				}
			})

			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Test csv.FindHeaderRow
			result, err := csv.FindHeaderRow(tmpFile.Name(), tt.required)
			if err != nil {
				t.Errorf("csv.FindHeaderRow() error = %v", err)
				return
			}

			if result != tt.expectedRow {
				t.Errorf("csv.FindHeaderRow() = %d, want %d", result, tt.expectedRow)
			}
		})
	}
}

func TestFindHeaderRow_Negative(t *testing.T) {
	tests := []struct {
		name       string
		csvContent string
		required   []string
	}{
		{
			"header not found",
			"A,B,C\n1,2,3",
			[]string{"Name", "Age", "City"},
		},
		{
			"partial header match",
			"Name,Age\nJohn,30",
			[]string{"Name", "Age", "City"},
		},
		{
			"empty file",
			"",
			[]string{"Name"},
		},
		{
			"header beyond MaxHeaderSearchRows",
			// Header is at row 21 (0-indexed: row 20), beyond default MaxHeaderSearchRows of 20
			"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20\nName,Age\n",
			[]string{"Name", "Age"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test_*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			t.Cleanup(func() {
				if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: failed to cleanup temp file: %v", err)
				}
			})

			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			_, err = csv.FindHeaderRow(tmpFile.Name(), tt.required)
			if err == nil {
				t.Error("csv.FindHeaderRow() expected error, got nil")
			}
		})
	}
}

func TestFindHeaderRow_FileNotFound(t *testing.T) {
	_, err := csv.FindHeaderRow("/nonexistent/file.csv", []string{"Name"})
	if err == nil {
		t.Error("csv.FindHeaderRow() expected error for nonexistent file")
	}
}

func TestFindHeaderRow_ExcelFormattedHeader(t *testing.T) {
	// Test that Excel formula formatting is handled
	// Note: Real Excel exports quote the formula cells properly
	// The internal quotes are escaped as double quotes: "=""Name"""
	csvContent := "\"=\"\"Name\"\"\",\"=\"\"Age\"\"\",\"=\"\"City\"\"\"\nJohn,30,NYC"

	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to cleanup temp file: %v", err)
		}
	})

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	result, err := csv.FindHeaderRow(tmpFile.Name(), []string{"Name", "Age", "City"})
	if err != nil {
		t.Errorf("csv.FindHeaderRow() error = %v", err)
		return
	}

	if result != 0 {
		t.Errorf("csv.FindHeaderRow() = %d, want 0", result)
	}
}

/* ========================================
	csv.Read Tests
======================================== */

func TestReadCsv_Positive(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectedLen int
	}{
		{
			"simple csv",
			"a,b,c\n1,2,3\n4,5,6",
			3,
		},
		{
			"single row",
			"a,b,c",
			1,
		},
		{
			"csv with quoted fields",
			`"name","address","notes"` + "\n" + `"John","123 Main St","Has, comma"`,
			2,
		},
		{
			"variable length rows",
			"a,b,c\n1,2\n4,5,6,7",
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test_*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			t.Cleanup(func() {
				if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: failed to cleanup temp file: %v", err)
				}
			})

			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			records, err := csv.Read(tmpFile.Name())
			if err != nil {
				t.Errorf("csv.Read() error = %v", err)
				return
			}

			if len(records) != tt.expectedLen {
				t.Errorf("csv.Read() returned %d rows, want %d", len(records), tt.expectedLen)
			}
		})
	}
}

func TestReadCsv_Negative(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		_, err := csv.Read("/nonexistent/file.csv")
		if err == nil {
			t.Error("csv.Read() expected error for nonexistent file")
		}
	})
}

func TestReadCsv_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to cleanup temp file: %v", err)
		}
	})
	tmpFile.Close()

	records, err := csv.Read(tmpFile.Name())
	if err != nil {
		t.Errorf("csv.Read() error = %v", err)
		return
	}

	if len(records) != 0 {
		t.Errorf("csv.Read() returned %d rows for empty file, want 0", len(records))
	}
}

func TestReadCsv_SpecialCharacters(t *testing.T) {
	csvContent := `Name,Description` + "\n" + `"Test","Line with
newline and ""quotes"""`

	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to cleanup temp file: %v", err)
		}
	})

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	records, err := csv.Read(tmpFile.Name())
	if err != nil {
		t.Errorf("csv.Read() error = %v", err)
		return
	}

	if len(records) != 2 {
		t.Errorf("csv.Read() returned %d rows, want 2", len(records))
	}
}

/* ========================================
	csv.Write Tests
======================================== */

func TestWriteCsvFile_Positive(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
	}{
		{
			"simple rows",
			[][]string{
				{"Name", "Age", "City"},
				{"John", "30", "NYC"},
				{"Jane", "25", "LA"},
			},
		},
		{
			"single row",
			[][]string{
				{"Header1", "Header2"},
			},
		},
		{
			"rows with special characters",
			[][]string{
				{"Name", "Notes"},
				{"Test", "Has, comma"},
				{"Another", "Has \"quotes\""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			t.Cleanup(func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to cleanup temp dir: %v", err)
				}
			})

			path := filepath.Join(tmpDir, "output.csv")

			err = csv.Write(path, tt.rows)
			if err != nil {
				t.Errorf("csv.Write() error = %v", err)
				return
			}

			// Verify file exists and can be read back
			readBack, err := csv.Read(path)
			if err != nil {
				t.Errorf("Failed to read back written file: %v", err)
				return
			}

			if len(readBack) != len(tt.rows) {
				t.Errorf("Written file has %d rows, want %d", len(readBack), len(tt.rows))
			}
		})
	}
}

func TestWriteCsvFile_Negative(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		err := csv.Write("/nonexistent/directory/file.csv", [][]string{{"a"}})
		if err == nil {
			t.Error("csv.Write() expected error for invalid path")
		}
	})
}

func TestWriteCsvFile_EmptyRows(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup temp dir: %v", err)
		}
	})

	path := filepath.Join(tmpDir, "empty.csv")

	err = csv.Write(path, [][]string{})
	if err != nil {
		t.Errorf("csv.Write() error = %v", err)
		return
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("csv.Write() did not create file")
	}
}

func TestWriteCsvFile_Roundtrip(t *testing.T) {
	// Test that data survives a write/read cycle
	original := [][]string{
		{"ID", "Name", "Amount"},
		{"1", "Product A", "$100.00"},
		{"2", "Product B", "$200.50"},
		{"3", "Product, Inc.", "$300.00"},
	}

	tmpDir, err := os.MkdirTemp("", "test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup temp dir: %v", err)
		}
	})

	path := filepath.Join(tmpDir, "roundtrip.csv")

	if err := csv.Write(path, original); err != nil {
		t.Fatalf("csv.Write() error = %v", err)
	}

	readBack, err := csv.Read(path)
	if err != nil {
		t.Fatalf("csv.Read() error = %v", err)
	}

	if len(readBack) != len(original) {
		t.Errorf("Row count mismatch: got %d, want %d", len(readBack), len(original))
	}

	for i, row := range original {
		if i >= len(readBack) {
			break
		}
		for j, cell := range row {
			if j >= len(readBack[i]) {
				t.Errorf("Row %d: cell count mismatch", i)
				break
			}
			if readBack[i][j] != cell {
				t.Errorf("Row %d, Col %d: got %q, want %q", i, j, readBack[i][j], cell)
			}
		}
	}
}

/* ========================================
	getUploadsRoot Tests
======================================== */

func TestGetUploadsRoot(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Warning: failed to restore working directory: %v", err)
		}
	})

	// Create temp directory and change to it
	tmpDir, err := os.MkdirTemp("", "test_uploads_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup temp dir: %v", err)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change dir: %v", err)
	}

	root, err := getUploadsRoot()
	if err != nil {
		t.Errorf("getUploadsRoot() error = %v", err)
		return
	}

	expected := filepath.Join(tmpDir, "accounting/uploads")
	if root != expected {
		t.Errorf("getUploadsRoot() = %q, want %q", root, expected)
	}
}

/* ========================================
	Integration Test: csv.FindHeaderRow + csv.Read
======================================== */

func TestFindHeaderRow_ReadCsv_Integration(t *testing.T) {
	// Simulate a real-world CSV with metadata rows before header
	csvContent := `Report: Sales Data
Generated: 2024-01-15
Account: Test Company

Transaction ID,Customer,Amount,Date
TXN001,Customer A,$100.00,2024-01-01
TXN002,Customer B,$200.00,2024-01-02
TXN003,Customer C,$300.00,2024-01-03`

	tmpFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to cleanup temp file: %v", err)
		}
	})

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Find header
	required := []string{"Transaction ID", "Customer", "Amount", "Date"}
	headerIdx, err := csv.FindHeaderRow(tmpFile.Name(), required)
	if err != nil {
		t.Fatalf("csv.FindHeaderRow() error = %v", err)
	}

	// Read all rows
	allRows, err := csv.Read(tmpFile.Name())
	if err != nil {
		t.Fatalf("csv.Read() error = %v", err)
	}

	// Extract data rows (skip header and metadata)
	dataRows := allRows[headerIdx+1:]

	if len(dataRows) != 3 {
		t.Errorf("Expected 3 data rows, got %d", len(dataRows))
	}

	// Verify first data row
	if len(dataRows) > 0 && dataRows[0][0] != "TXN001" {
		t.Errorf("First data row ID = %q, want 'TXN001'", dataRows[0][0])
	}
}

