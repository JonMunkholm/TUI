package handler

import (
	"testing"

	"github.com/JonMunkholm/TUI/internal/schema"
)

// sfdcTestHeader returns the header row matching SfdcFieldSpecs order
func sfdcTestHeader() []string {
	headers := make([]string, len(schema.SfdcFieldSpecs))
	for i, spec := range schema.SfdcFieldSpecs {
		headers[i] = spec.Name
	}
	return headers
}

// sfdcTestHeaderIndex returns a pre-computed HeaderIndex for tests
func sfdcTestHeaderIndex() HeaderIndex {
	return MakeHeaderIndex(sfdcTestHeader())
}

// sfdcValidRow returns a row with valid values for all sfdc fields
func sfdcValidRow() []string {
	return []string{
		"0060R00001TEST01",   // 0: Opportunity ID Casesafe
		"00k0R00001TEST01",   // 1: Opportunity Product Casesafe ID
		"Test Opportunity",   // 2: Opportunity Name
		"Test Account",       // 3: Account Name
		"2024-01-15",         // 4: Close Date
		"2024-01-10",         // 5: Booked Date
		"FY24-Q1",            // 6: Fiscal Period
		"Annual",             // 7: Payment Schedule
		"Net 30",             // 8: Payment Due
		"2024-01-01",         // 9: Contract Start Date
		"2024-12-31",         // 10: Contract End Date
		"12",                 // 11: Term in Months_deprecated
		"Test Product",       // 12: Product Name
		"Cloud",              // 13: Deployment Type
		"10000",              // 14: Amount
		"10",                 // 15: Quantity
		"1200",               // 16: List Price
		"1000",               // 17: Sales Price
		"10000",              // 18: Total Price
		"2024-01-01",         // 19: Start Date
		"2024-12-31",         // 20: End Date
		"12",                 // 21: Term in Months
		"PROD-001",           // 22: Product Code
		"9000",               // 23: Total Amount Due - Customer
		"1000",               // 24: Total Amount Due - Partner
		"true",               // 25: Active Product
	}
}

/* ========================================
	BuildSfdcOppLineItemParams Tests
======================================== */

func TestBuildSfdcOppLineItemParams_Positive(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	tests := []struct {
		name string
		row  []string
	}{
		{
			"complete row with all fields",
			[]string{
				"0060R00001ABC123", "00k0R00001XYZ789", "Enterprise Deal - ACME Corp",
				"ACME Corporation", "2024-01-15", "2024-01-10",
				"FY24-Q1", "Annual", "Net 30",
				"2024-01-01", "2024-12-31", "12",
				"Enterprise License", "Cloud",
				"100000", "10", "12000", "10000", "100000",
				"2024-01-01", "2024-12-31", "12",
				"ENT-LIC-001", "90000", "10000", "true",
			},
		},
		{
			"row with minimal required fields",
			[]string{
				"0060R00001DEF456", "00k0R00001UVW321", "Small Deal",
				"Small Co", "2024-03-01", "2024-03-01",
				"FY24-Q1", "One-time", "Net 30",
				"2024-03-01", "2024-03-31", "1",
				"Basic License", "SaaS",
				"1000", "1", "1000", "1000", "1000",
				"2024-03-01", "2024-03-31", "1",
				"BASIC-001", "1000", "0", "false",
			},
		},
		{
			"row with currency formatting",
			[]string{
				"0060R00001GHI789", "00k0R00001RST654", "Mid-Market Deal",
				"Mid Corp", "01/15/2024", "01/10/2024",
				"FY24-Q1", "Monthly", "Net 15",
				"01/01/2024", "12/31/2024", "12",
				"Pro License", "Hybrid",
				"$50,000.00", "5", "$12,000.00", "$10,000.00", "$50,000.00",
				"01/01/2024", "12/31/2024", "12",
				"PRO-LIC-001", "$45,000.00", "$5,000.00", "yes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := s.BuildSfdcOppLineItemParams(tt.row, headerIdx)
			if err != nil {
				t.Errorf("BuildSfdcOppLineItemParams() error = %v", err)
				return
			}

			// Verify OpportunityID is set
			if !params.OpportunityID.Valid {
				t.Errorf("OpportunityID should be valid")
			}
		})
	}
}

func TestBuildSfdcOppLineItemParams_Negative(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

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
			"row with only 10 fields - should error",
			[]string{
				"ID1", "ID2", "Name", "Account",
				"2024-01-15", "2024-01-10", "Q1", "Annual",
				"Net 30", "2024-01-01",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.BuildSfdcOppLineItemParams(tt.row, headerIdx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but didn't get one")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBuildSfdcOppLineItemParams_BooleanFields(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	tests := []struct {
		name           string
		activeProduct  string
		expectedBool   bool
		expectedValid  bool
		expectError    bool
	}{
		{"true lowercase", "true", true, true, false},
		{"TRUE uppercase", "TRUE", true, true, false},
		{"yes", "yes", true, true, false},
		{"1", "1", true, true, false},
		{"false lowercase", "false", false, true, false},
		{"FALSE uppercase", "FALSE", false, true, false},
		{"no", "no", false, true, false},
		{"0", "0", false, true, false},
		{"empty string", "", false, false, true},     // Required - expect error
		{"invalid string", "maybe", false, false, true}, // Invalid - expect error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := sfdcValidRow()
			row[25] = tt.activeProduct

			params, err := s.BuildSfdcOppLineItemParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for value %q, got none", tt.activeProduct)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildSfdcOppLineItemParams() error = %v", err)
				return
			}

			if params.ActiveProduct.Valid != tt.expectedValid {
				t.Errorf("ActiveProduct.Valid = %v, want %v", params.ActiveProduct.Valid, tt.expectedValid)
			}

			if params.ActiveProduct.Valid && params.ActiveProduct.Bool != tt.expectedBool {
				t.Errorf("ActiveProduct.Bool = %v, want %v", params.ActiveProduct.Bool, tt.expectedBool)
			}
		})
	}
}

func TestBuildSfdcOppLineItemParams_DateFields(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	tests := []struct {
		name        string
		closeDate   string
		expectValid bool
		expectError bool
	}{
		{"ISO format", "2024-01-15", true, false},
		{"US format with slashes", "01/15/2024", true, false},
		{"US format short year", "01/15/24", true, false},
		{"text month", "Jan 15, 2024", true, false},
		{"empty", "", false, true},       // Required - expect error
		{"invalid", "not-a-date", false, true}, // Invalid - expect error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := sfdcValidRow()
			row[4] = tt.closeDate // CloseDate field

			params, err := s.BuildSfdcOppLineItemParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for date value %q, got none", tt.closeDate)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildSfdcOppLineItemParams() error = %v", err)
				return
			}

			if params.CloseDate.Valid != tt.expectValid {
				t.Errorf("CloseDate.Valid = %v, want %v for input %q",
					params.CloseDate.Valid, tt.expectValid, tt.closeDate)
			}
		})
	}
}

func TestBuildSfdcOppLineItemParams_NumericFields(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	tests := []struct {
		name        string
		amount      string
		expectValid bool
		expectError bool
	}{
		{"integer", "100000", true, false},
		{"decimal", "100000.50", true, false},
		{"with dollar sign", "$100,000.00", true, false},
		{"negative", "-5000", true, false},
		{"accounting negative", "(5000.00)", true, false},
		{"empty", "", false, true},         // Required - expect error
		{"invalid text", "one hundred", false, true}, // Invalid - expect error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := sfdcValidRow()
			row[14] = tt.amount // Amount field

			params, err := s.BuildSfdcOppLineItemParams(row, headerIdx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for amount value %q, got none", tt.amount)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildSfdcOppLineItemParams() error = %v", err)
				return
			}

			if params.Amount.Valid != tt.expectValid {
				t.Errorf("Amount.Valid = %v, want %v for input %q",
					params.Amount.Valid, tt.expectValid, tt.amount)
			}
		})
	}
}

/* ========================================
	False Positive/Negative Detection
======================================== */

func TestBuildSfdcOppLineItemParams_FalsePositives(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	tests := []struct {
		name        string
		fieldIndex  int
		fieldValue  string
		fieldName   string
		description string
	}{
		{
			"phone number should not be valid numeric",
			14, // Amount
			"555-123-4567",
			"Amount",
			"Phone numbers should not parse as valid numerics",
		},
		{
			"IP address should not be valid numeric",
			14,
			"192.168.1.1",
			"Amount",
			"IP addresses have multiple dots and should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := sfdcValidRow()
			row[tt.fieldIndex] = tt.fieldValue

			// These invalid values should cause validation errors for required numeric fields
			_, err := s.BuildSfdcOppLineItemParams(row, headerIdx)
			if err == nil {
				t.Errorf("Expected validation error for invalid numeric: %s", tt.description)
			}
		})
	}
}

/* ========================================
	Real World Data Tests
======================================== */

func TestBuildSfdcOppLineItemParams_RealWorldData(t *testing.T) {
	s := &SfdcUpload{}
	headerIdx := sfdcTestHeaderIndex()

	realWorldRows := [][]string{
		// Standard Salesforce export
		{
			"0060R00001ABCDEF", "00k0R00001GHIJKL", "Enterprise Agreement - Tech Corp",
			"Technology Corporation", "2024-03-15", "2024-03-10",
			"FY24-Q1", "Annual Upfront", "Due on Receipt",
			"2024-04-01", "2025-03-31", "12",
			"Enterprise Platform License", "SaaS",
			"250000", "50", "6000", "5000", "250000",
			"2024-04-01", "2025-03-31", "12",
			"EPL-2024", "225000", "25000", "TRUE",
		},
		// Export with Excel formatting
		{
			`="0060R00001MNOPQR"`, `="00k0R00001STUVWX"`, `="Renewal - BigCo Inc"`,
			`="BigCo Incorporated"`, "01/31/2024", "01/25/2024",
			"FY24-Q4", "Quarterly", "Net 30",
			"02/01/2024", "01/31/2025", "12",
			"Standard License", "On-Premise",
			`="$75,000.00"`, "25", `="$3,600.00"`, `="$3,000.00"`, `="$75,000.00"`,
			"02/01/2024", "01/31/2025", "12",
			"STD-2024", `="$67,500.00"`, `="$7,500.00"`, "Yes",
		},
	}

	for i, row := range realWorldRows {
		t.Run("real_world_sfdc_"+string(rune('A'+i)), func(t *testing.T) {
			params, err := s.BuildSfdcOppLineItemParams(row, headerIdx)
			if err != nil {
				t.Errorf("Failed to build params: %v", err)
				return
			}

			// Key fields should be valid
			if !params.OpportunityID.Valid {
				t.Error("OpportunityID should be valid")
			}
			if !params.OpportunityName.Valid {
				t.Error("OpportunityName should be valid")
			}
			if !params.Amount.Valid {
				t.Error("Amount should be valid")
			}
			if !params.ActiveProduct.Valid {
				t.Error("ActiveProduct should be valid")
			}
		})
	}
}
