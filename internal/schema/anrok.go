package schema

// AnrokFieldSpecs defines the expected CSV columns for Anrok tax transaction reports.
var AnrokFieldSpecs = []FieldSpec{
	{Name: "Transaction ID", Type: FieldText, Required: true},
	{Name: "Customer ID", Type: FieldText, Required: true},
	{Name: "Customer name", Type: FieldText, Required: true},
	{Name: "Overall VAT ID validation status", Type: FieldText, Required: true},
	{Name: "Valid VAT IDs", Type: FieldText, Required: false},
	{Name: "Other VAT IDs", Type: FieldText, Required: false},
	{Name: "Invoice date", Type: FieldDate, Required: true},
	{Name: "Tax date", Type: FieldDate, Required: true},
	{Name: "Transaction currency", Type: FieldText, Required: true},
	{Name: "Sales amount", Type: FieldNumeric, Required: true},
	{Name: "Exempt reasons", Type: FieldText, Required: false},
	{Name: "Tax amount", Type: FieldNumeric, Required: true},
	{Name: "Invoice amount", Type: FieldNumeric, Required: true},
	{Name: "Void", Type: FieldBool, Required: true},
	{Name: "Customer address line 1", Type: FieldText, Required: true},
	{Name: "Customer address city", Type: FieldText, Required: true},
	{Name: "Customer address region", Type: FieldText, Required: true, Normalizer: NormalizeUsState},
	{Name: "Customer address postal code", Type: FieldText, Required: true},
	{Name: "Customer address country", Type: FieldText, Required: true},
	{Name: "Customer country code", Type: FieldText, Required: true},
	{Name: "Jurisdictions", Type: FieldText, Required: true},
	{Name: "Jurisdictions IDs", Type: FieldText, Required: true},
	{Name: "Return IDs", Type: FieldText, Required: true},
}
