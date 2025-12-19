// Package schema defines field specifications for CSV data validation.
// Each data source (NS, SFDC, Anrok) has its own field specs that define
// the expected CSV structure and validation rules.
package schema

// FieldType represents the expected data type for a CSV field.
type FieldType int

const (
	FieldText FieldType = iota
	FieldEnum
	FieldDate
	FieldNumeric
	FieldBool
)

// FieldSpec defines validation rules for a single CSV column.
type FieldSpec struct {
	Name       string            // Column header name (must match CSV exactly)
	Type       FieldType         // Expected data type
	Required   bool              // Column must exist in CSV header
	AllowEmpty bool              // If true, empty values are allowed even when Required
	EnumValues []string          // Valid values for FieldEnum type
	Normalizer func(string) string // Optional transformation function
}
