package handler

import (
	"testing"

	"github.com/JonMunkholm/TUI/internal/schema"
)

// anrokTestHeader returns the header row matching AnrokFieldSpecs order
func anrokTestHeader() []string {
	headers := make([]string, len(schema.AnrokFieldSpecs))
	for i, spec := range schema.AnrokFieldSpecs {
		headers[i] = spec.Name
	}
	return headers
}

// anrokTestHeaderIndex returns a pre-computed HeaderIndex for tests
func anrokTestHeaderIndex() HeaderIndex {
	return MakeHeaderIndex(anrokTestHeader())
}

// anrokValidRow returns a row with valid values for all anrok fields
// Tests can copy and modify specific fields as needed
func anrokValidRow() []string {
	return []string{
		"TXN-TEST",       // 0: Transaction ID
		"CUST-TEST",      // 1: Customer ID
		"Test Customer",  // 2: Customer name
		"Validated",      // 3: Overall VAT ID validation status
		"VAT123",         // 4: Valid VAT IDs
		"",               // 5: Other VAT IDs
		"2024-01-15",     // 6: Invoice date
		"2024-01-15",     // 7: Tax date
		"USD",            // 8: Transaction currency
		"1000.00",        // 9: Sales amount
		"",               // 10: Exempt reasons
		"80.00",          // 11: Tax amount
		"1080.00",        // 12: Invoice amount
		"false",          // 13: Void
		"123 Main St",    // 14: Customer address line 1
		"Test City",      // 15: Customer address city
		"CA",             // 16: Customer address region
		"94105",          // 17: Customer address postal code
		"United States",  // 18: Customer address country
		"US",             // 19: Customer country code
		"California",     // 20: Jurisdictions
		"CA-001",         // 21: Jurisdictions IDs
		"RET-001",        // 22: Return IDs
	}
}

/* ========================================
	BuildAnrokTransactionParams Tests
======================================== */

func TestBuildAnrokTransactionParams_Positive(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name string
		row  []string
	}{
		{
			"complete US transaction",
			[]string{
				"TXN-001", "CUST-001", "ACME Corporation",
				"Validated", "US123456789", "",
				"2024-01-15", "2024-01-15", "USD",
				"1000.00", "", "80.00", "1080.00",
				"false", "123 Main St", "San Francisco",
				"CA", "94105", "United States", "US",
				"California", "CA-001", "RET-001",
			},
		},
		{
			"EU transaction with VAT",
			[]string{
				"TXN-002", "CUST-002", "Euro GmbH",
				"Validated", "DE123456789", "FR987654321",
				"2024-02-01", "2024-02-01", "EUR",
				"5000.00", "Reverse Charge", "0.00", "5000.00",
				"false", "Hauptstraße 1", "Berlin",
				"BE", "10115", "Germany", "DE",
				"Germany", "DE-001", "RET-002",
			},
		},
		{
			"voided transaction",
			[]string{
				"TXN-003", "CUST-003", "Voided Corp",
				"Validated", "", "",
				"2024-01-01", "2024-01-01", "USD",
				"500.00", "", "40.00", "540.00",
				"true", "456 Oak Ave", "Los Angeles",
				"CA", "90001", "United States", "US",
				"California", "CA-001", "RET-003",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := a.BuildAnrokTransactionParams(tt.row, headerIdx)
			if err != nil {
				t.Errorf("BuildAnrokTransactionParams() error = %v", err)
				return
			}

			// Verify TransactionID is set
			if !params.TransactionID.Valid {
				t.Error("TransactionID should be valid")
			}
		})
	}
}

func TestBuildAnrokTransactionParams_Negative(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name        string
		row         []string
		expectError bool
	}{
		{
			"empty row - should error",
			[]string{},
			true,
		},
		{
			"partial row with only 10 fields - should error",
			[]string{
				"TXN-001", "CUST-001", "Customer",
				"Validated", "VAT123", "",
				"2024-01-15", "2024-01-15", "USD",
				"1000.00",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := a.BuildAnrokTransactionParams(tt.row, headerIdx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but didn't get one")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBuildAnrokTransactionParams_StateNormalization(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name          string
		stateInput    string
		expectedState string
	}{
		{"lowercase state code", "ca", "CA"},
		{"uppercase state code", "CA", "CA"},
		{"full state name", "california", "CA"},
		{"full state name mixed case", "California", "CA"},
		{"two word state name", "new york", "NY"},
		{"foreign region passes through", "Bavaria", "Bavaria"},
		// Note: empty state test removed - Customer address region is a required field
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := anrokValidRow()
			row[16] = tt.stateInput // CustomerAddressRegion field

			params, err := a.BuildAnrokTransactionParams(row, headerIdx)
			if err != nil {
				t.Errorf("BuildAnrokTransactionParams() error = %v", err)
				return
			}

			if params.CustomerAddressRegion.String != tt.expectedState {
				t.Errorf("State = %q, want %q", params.CustomerAddressRegion.String, tt.expectedState)
			}
		})
	}
}

func TestBuildAnrokTransactionParams_VoidField(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name          string
		voidValue     string
		expectedBool  bool
		expectedValid bool
		expectError   bool
	}{
		{"true lowercase", "true", true, true, false},
		{"TRUE uppercase", "TRUE", true, true, false},
		{"false lowercase", "false", false, true, false},
		{"FALSE uppercase", "FALSE", false, true, false},
		{"yes", "yes", true, true, false},
		{"no", "no", false, true, false},
		{"1", "1", true, true, false},
		{"0", "0", false, true, false},
		{"empty", "", false, false, true},    // Required field - expect error
		{"invalid", "maybe", false, false, true}, // Invalid bool - expect error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := anrokValidRow()
			row[13] = tt.voidValue // Void field

			params, err := a.BuildAnrokTransactionParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for void value %q, got none", tt.voidValue)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildAnrokTransactionParams() error = %v", err)
				return
			}

			if params.Void.Valid != tt.expectedValid {
				t.Errorf("Void.Valid = %v, want %v", params.Void.Valid, tt.expectedValid)
			}

			if params.Void.Valid && params.Void.Bool != tt.expectedBool {
				t.Errorf("Void.Bool = %v, want %v", params.Void.Bool, tt.expectedBool)
			}
		})
	}
}

func TestBuildAnrokTransactionParams_NumericFields(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name        string
		fieldIndex  int
		fieldValue  string
		fieldName   string
		expectValid bool
		expectError bool
	}{
		{"sales amount integer", 9, "1000", "SalesAmount", true, false},
		{"sales amount decimal", 9, "1000.50", "SalesAmount", true, false},
		{"sales amount with currency", 9, "$1,000.50", "SalesAmount", true, false},
		{"sales amount negative", 9, "-500.00", "SalesAmount", true, false},
		{"sales amount accounting", 9, "(500.00)", "SalesAmount", true, false},
		{"sales amount empty", 9, "", "SalesAmount", false, true},      // Required - expect error
		{"sales amount invalid", 9, "not-a-number", "SalesAmount", false, true}, // Invalid - expect error
		{"tax amount", 11, "80.25", "TaxAmount", true, false},
		{"invoice amount", 12, "1080.25", "InvoiceAmount", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := anrokValidRow()
			row[tt.fieldIndex] = tt.fieldValue

			params, err := a.BuildAnrokTransactionParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s value %q, got none", tt.fieldName, tt.fieldValue)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildAnrokTransactionParams() error = %v", err)
				return
			}

			var valid bool
			switch tt.fieldIndex {
			case 9:
				valid = params.SalesAmount.Valid
			case 11:
				valid = params.TaxAmount.Valid
			case 12:
				valid = params.InvoiceAmount.Valid
			}

			if valid != tt.expectValid {
				t.Errorf("%s.Valid = %v, want %v for input %q",
					tt.fieldName, valid, tt.expectValid, tt.fieldValue)
			}
		})
	}
}

func TestBuildAnrokTransactionParams_DateFields(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name        string
		dateValue   string
		expectValid bool
		expectError bool
	}{
		{"ISO format", "2024-01-15", true, false},
		{"US format", "01/15/2024", true, false},
		{"short year", "01/15/24", true, false},
		{"text month", "Jan 15, 2024", true, false},
		{"empty", "", false, true},       // Required - expect error
		{"invalid", "not-a-date", false, true}, // Invalid - expect error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := anrokValidRow()
			row[6] = tt.dateValue // InvoiceDate field

			params, err := a.BuildAnrokTransactionParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for date value %q, got none", tt.dateValue)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildAnrokTransactionParams() error = %v", err)
				return
			}

			if params.InvoiceDate.Valid != tt.expectValid {
				t.Errorf("InvoiceDate.Valid = %v, want %v for input %q",
					params.InvoiceDate.Valid, tt.expectValid, tt.dateValue)
			}
		})
	}
}

/* ========================================
	False Positive/Negative Detection
======================================== */

func TestBuildAnrokTransactionParams_FalsePositives(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	tests := []struct {
		name        string
		fieldIndex  int
		fieldValue  string
		description string
	}{
		{
			"version number should not be valid amount",
			9, // SalesAmount
			"1.2.3",
			"Version numbers have multiple dots",
		},
		{
			"phone should not be valid amount",
			9,
			"555-123-4567",
			"Phone numbers are not numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := anrokValidRow()
			row[tt.fieldIndex] = tt.fieldValue

			// These invalid values should cause validation errors for required numeric fields
			_, err := a.BuildAnrokTransactionParams(row, headerIdx)
			if err == nil {
				t.Errorf("Expected validation error for invalid numeric: %s", tt.description)
			}
		})
	}
}

/* ========================================
	Real World Data Tests
======================================== */

func TestBuildAnrokTransactionParams_RealWorldData(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	realWorldRows := [][]string{
		// Standard Anrok export - US transaction
		{
			"inv_01ABC123DEF456", "cus_01GHI789JKL012",
			"TechStartup Inc.",
			"validated", "US12-3456789", "",
			"2024-01-15", "2024-01-15", "USD",
			"9999.00", "", "824.92", "10823.92",
			"false",
			"100 Innovation Way", "San Francisco",
			"California", "94107", "United States", "US",
			"California State Tax|San Francisco County Tax",
			"CA-STATE-001|CA-SF-COUNTY-001",
			"RET-2024-01-CA",
		},
		// International transaction with VAT exemption
		{
			"inv_02MNO345PQR678", "cus_02STU901VWX234",
			"EuroTech GmbH",
			"validated", "DE123456789", "",
			"2024-02-01", "2024-02-01", "EUR",
			"25000.00", "B2B Reverse Charge", "0.00", "25000.00",
			"false",
			"Techstraße 42", "Munich",
			"BY", "80331", "Germany", "DE",
			"Germany VAT", "DE-VAT-001", "RET-2024-02-DE",
		},
		// Voided transaction
		{
			"inv_03YZA567BCD890", "cus_03EFG123HIJ456",
			"Cancelled Corp",
			"Validated", "", "",
			"2024-01-01", "2024-01-01", "USD",
			"500.00", "", "41.25", "541.25",
			"TRUE",
			"999 Void Street", "Austin",
			"TX", "78701", "United States", "US",
			"Texas State Tax", "TX-STATE-001", "RET-2024-01-TX",
		},
	}

	for i, row := range realWorldRows {
		t.Run("real_world_anrok_"+string(rune('A'+i)), func(t *testing.T) {
			params, err := a.BuildAnrokTransactionParams(row, headerIdx)
			if err != nil {
				t.Errorf("Failed to build params: %v", err)
				return
			}

			// Key fields should be valid
			if !params.TransactionID.Valid {
				t.Error("TransactionID should be valid")
			}
			if !params.CustomerID.Valid {
				t.Error("CustomerID should be valid")
			}
			if !params.SalesAmount.Valid {
				t.Error("SalesAmount should be valid")
			}

			// State normalization should work
			if row[16] == "California" && params.CustomerAddressRegion.String != "CA" {
				t.Errorf("State should be normalized to CA, got %q",
					params.CustomerAddressRegion.String)
			}
			if row[16] == "TX" && params.CustomerAddressRegion.String != "TX" {
				t.Errorf("State code should remain TX, got %q",
					params.CustomerAddressRegion.String)
			}
		})
	}
}

/* ========================================
	Edge Case Tests
======================================== */

func TestBuildAnrokTransactionParams_ExcelFormatting(t *testing.T) {
	a := &AnrokUpload{}
	headerIdx := anrokTestHeaderIndex()

	row := []string{
		`="inv_excel_001"`, `="cus_excel_001"`,
		`="Excel Customer"`,
		"validated", `="VAT123456"`, "",
		"01/15/2024", "01/15/2024", "USD",
		`="$1,000.00"`, "", `="$82.50"`, `="$1,082.50"`,
		"false",
		`="123 Excel St"`, `="Test City"`,
		`="CA"`, `="94105"`, `="United States"`, `="US"`,
		"Excel Jurisdiction", "JUR-001", "RET-001",
	}

	params, err := a.BuildAnrokTransactionParams(row, headerIdx)
	if err != nil {
		t.Errorf("BuildAnrokTransactionParams() error = %v", err)
		return
	}

	// Verify Excel formatting was stripped
	if params.TransactionID.String != "inv_excel_001" {
		t.Errorf("TransactionID = %q, want 'inv_excel_001'", params.TransactionID.String)
	}

	if params.CustomerName.String != "Excel Customer" {
		t.Errorf("CustomerName = %q, want 'Excel Customer'", params.CustomerName.String)
	}
}
