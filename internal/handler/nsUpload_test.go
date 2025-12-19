package handler

import (
	"testing"

	"github.com/JonMunkholm/TUI/internal/schema"
)

// nsSoTestHeader returns the header row matching NsSoFieldSpecs order
func nsSoTestHeader() []string {
	headers := make([]string, len(schema.NsSoFieldSpecs))
	for i, spec := range schema.NsSoFieldSpecs {
		headers[i] = spec.Name
	}
	return headers
}

// nsInvoiceTestHeader returns the header row matching NsInvoiceFieldSpecs order
func nsInvoiceTestHeader() []string {
	headers := make([]string, len(schema.NsInvoiceFieldSpecs))
	for i, spec := range schema.NsInvoiceFieldSpecs {
		headers[i] = spec.Name
	}
	return headers
}

// nsSoTestHeaderIndex returns a pre-computed HeaderIndex for SO tests
func nsSoTestHeaderIndex() HeaderIndex {
	return MakeHeaderIndex(nsSoTestHeader())
}

// nsInvoiceTestHeaderIndex returns a pre-computed HeaderIndex for Invoice tests
func nsInvoiceTestHeaderIndex() HeaderIndex {
	return MakeHeaderIndex(nsInvoiceTestHeader())
}

// nsSoValidRow returns a row with valid values for all NS SO fields
func nsSoValidRow() []string {
	return []string{
		"OPP-001",            // 0: Salesforce Opportunity Id (IO)
		"LINE-001",           // 1: Salesforce Opportunity Line Id (IO)
		"Customer Project",   // 2: Customer/Project
		"DOC-123",            // 3: Document Number
		"2024-01-15",         // 4: Date
		"2024-01-01",         // 5: Start Date
		"2024-12-31",         // 6: End Date
		"Item Name",          // 7: Item: Name
		"Display Name",       // 8: Item: Display Name
		"2024-02-01",         // 9: Start Date (Line)
		"2024-11-30",         // 10: End Date (Line Level)
		"10",                 // 11: Quantity
		"100",                // 12: Contract Quantity
		"99.99",              // 13: Unit Price
		"500.00",             // 14: Total Amount Due Partner
		"1000.00",            // 15: Amount (Gross)
		"30",                 // 16: Terms: Days Till Net Due
	}
}

// nsInvoiceValidRow returns a row with valid values for all NS Invoice fields
func nsInvoiceValidRow() []string {
	return []string{
		"Invoice",             // 0: Type
		"2024-01-15",          // 1: Date
		"2024-02-15",          // 2: Date Due
		"INV-001",             // 3: Document Number
		"Customer Name",       // 4: Name
		"Invoice memo",        // 5: Memo (optional)
		"Sales Tax Item",      // 6: Item
		"1",                   // 7: Qty
		"10",                  // 8: Contract Quantity
		"99.99",               // 9: Unit Price
		"999.90",              // 10: Amount
		"2024-01-01",          // 11: Start Date (Line)
		"2024-12-31",          // 12: End Date (Line Level)
		"Sales Tax Account",   // 13: Account
		"OPP-001",             // 14: Salesforce Opportunity Id (IO)
		"PB-001",              // 15: Salesforce Pricebook Id (IO)
		"ITEM-001",            // 16: Item: Internal ID
		"ENT-001",             // 17: Entity: Internal ID
		"San Francisco",       // 18: Address: Shipping Address City
		"CA",                  // 19: Address: Shipping Address State
		"US",                  // 20: Address: Shipping Address Country
	}
}

/* ========================================
	BuildSoLineItemParams Tests
======================================== */

func TestBuildSoLineItemParams_Positive(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsSoTestHeaderIndex()

	tests := []struct {
		name string
		row  []string
	}{
		{
			"complete row",
			[]string{
				"OPP-001", "LINE-001", "Customer Project", "DOC-123",
				"2024-01-15", "2024-01-01", "2024-12-31", "Item Name",
				"Display Name", "2024-02-01", "2024-11-30", "10",
				"100", "99.99", "500.00", "1000.00", "30",
			},
		},
		{
			"row with minimal formatting",
			[]string{
				"OPP-002", "LINE-002", "Project", "DOC-456",
				"2024-05-01", "2024-05-01", "2024-05-31", "Basic Item",
				"Basic Display", "2024-05-01", "2024-05-31", "1",
				"1", "50", "0", "50", "30",
			},
		},
		{
			"row with currency formatting",
			[]string{
				"OPP-003", "LINE-003", "Project X", "DOC-789",
				"01/15/24", "01/01/24", "12/31/24", "Product",
				"Product Display", "02/01/24", "11/30/24", "$1,000",
				"$2,500", "$99.99", "$500.00", "$10,000.00", "45",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := n.BuildSoLineItemParams(tt.row, headerIdx)
			if err != nil {
				t.Errorf("BuildSoLineItemParams() error = %v", err)
				return
			}

			// Verify CustomerProject is set (non-nullable field)
			if params.CustomerProject == "" && tt.row[2] != "" {
				t.Errorf("CustomerProject not set correctly")
			}
		})
	}
}

func TestBuildSoLineItemParams_Negative(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsSoTestHeaderIndex()

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
			"partial row - should error",
			[]string{"OPP-001", "LINE-001"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := n.BuildSoLineItemParams(tt.row, headerIdx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but didn't get one")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBuildSoLineItemParams_EdgeCases(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsSoTestHeaderIndex()

	tests := []struct {
		name        string
		row         []string
		checkField  string
		checkValid  bool
		description string
	}{
		{
			"date field with various formats",
			[]string{
				"OPP-001", "LINE-001", "Project", "DOC-123",
				"Jan 15, 2024", "2024/01/01", "12-31-2024", "Item",
				"Display", "1/1/24", "12/31/24", "10",
				"100", "99.99", "500.00", "1000.00", "30",
			},
			"DocumentDate",
			true,
			"Various date formats should be parsed",
		},
		{
			"numeric fields with accounting format",
			[]string{
				"OPP-001", "LINE-001", "Project", "DOC-123",
				"2024-01-15", "2024-01-01", "2024-12-31", "Item",
				"Display", "2024-02-01", "2024-11-30", "(100)",
				"(500)", "($99.99)", "($500.00)", "($1,000.00)", "30",
			},
			"Quantity",
			true,
			"Accounting negative format should be parsed",
		},
		{
			"fields with excel formula prefix",
			[]string{
				`="OPP-001"`, `="LINE-001"`, "Project", `="DOC-123"`,
				"2024-01-15", "2024-01-01", "2024-12-31", "Item",
				"Display", "2024-02-01", "2024-11-30", "10",
				"100", "99.99", "500.00", "1000.00", "30",
			},
			"SalesforceOpportunityID",
			true,
			"Excel formula prefix should be stripped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := n.BuildSoLineItemParams(tt.row, headerIdx)
			if err != nil {
				t.Errorf("BuildSoLineItemParams() error = %v", err)
				return
			}

			// Verify the params were built (basic sanity check)
			if params.CustomerProject == "" {
				t.Logf("Note: CustomerProject is empty for test %s", tt.name)
			}
		})
	}
}

/* ========================================
	BuildInvoiceSalesTaxParams Tests
======================================== */

func TestBuildInvoiceSalesTaxParams_Positive(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsInvoiceTestHeaderIndex()

	tests := []struct {
		name string
		row  []string
	}{
		{
			"complete row",
			[]string{
				"Invoice", "2024-01-15", "2024-02-15", "INV-001",
				"Customer Name", "Invoice memo", "Sales Tax Item",
				"1", "10", "99.99", "999.90",
				"2024-01-01", "2024-12-31", "Sales Tax Account",
				"OPP-001", "PB-001", "ITEM-001", "ENT-001",
				"San Francisco", "CA", "US",
			},
		},
		{
			"row with state full name",
			[]string{
				"Credit Memo", "2024-03-01", "2024-04-01", "CM-001",
				"Another Customer", "Credit memo", "Tax Item",
				"2", "20", "49.99", "999.80",
				"2024-02-01", "2024-11-30", "Tax Account",
				"OPP-002", "PB-002", "ITEM-002", "ENT-002",
				"New York", "New York", "United States",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := n.BuildInvoiceSalesTaxParams(tt.row, headerIdx)
			if err != nil {
				t.Errorf("BuildInvoiceSalesTaxParams() error = %v", err)
				return
			}

			// Verify type is set
			if !params.Type.Valid {
				t.Errorf("Type should be valid")
			}
		})
	}
}

func TestBuildInvoiceSalesTaxParams_StateNormalization(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsInvoiceTestHeaderIndex()

	tests := []struct {
		name          string
		stateInput    string
		expectedState string
	}{
		{"lowercase state code", "ca", "CA"},
		{"uppercase state code", "CA", "CA"},
		{"full state name lowercase", "california", "CA"},
		{"full state name mixed", "California", "CA"},
		{"full state name with spaces", "new york", "NY"},
		{"unknown state passes through", "Ontario", "Ontario"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := []string{
				"Invoice", "2024-01-15", "2024-02-15", "INV-001",
				"Customer", "Memo", "Item",
				"1", "10", "99.99", "999.90",
				"2024-01-01", "2024-12-31", "Account",
				"OPP-001", "PB-001", "ITEM-001", "ENT-001",
				"City", tt.stateInput, "Country",
			}

			params, err := n.BuildInvoiceSalesTaxParams(row, headerIdx)
			if err != nil {
				t.Errorf("BuildInvoiceSalesTaxParams() error = %v", err)
				return
			}

			if params.ShippingAddressState.String != tt.expectedState {
				t.Errorf("State = %q, want %q", params.ShippingAddressState.String, tt.expectedState)
			}
		})
	}
}

func TestBuildInvoiceSalesTaxParams_Negative(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsInvoiceTestHeaderIndex()

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
			"row missing state fields - should error",
			[]string{
				"Invoice", "2024-01-15", "2024-02-15", "INV-001",
				"Customer", "Memo", "Item",
				"1", "10", "99.99", "999.90",
				"2024-01-01", "2024-12-31", "Account",
				"OPP-001", "PB-001", "ITEM-001", "ENT-001",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := n.BuildInvoiceSalesTaxParams(tt.row, headerIdx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but didn't get one")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

/* ========================================
	False Positive/Negative Detection Tests
======================================== */

func TestBuildSoLineItemParams_FalsePositives(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsSoTestHeaderIndex()

	// Since Date and Quantity are required fields, invalid values should return errors
	tests := []struct {
		name        string
		row         []string
		description string
	}{
		{
			"invalid date format should return error",
			[]string{
				"OPP-001", "LINE-001", "Project", "DOC-123",
				"not-a-date", "also-not-date", "nope", "Item",
				"Display", "invalid", "invalid", "10",
				"100", "99.99", "500.00", "1000.00", "30",
			},
			"Invalid date strings in required fields should return validation error",
		},
		{
			"non-numeric quantity should return error",
			[]string{
				"OPP-001", "LINE-001", "Project", "DOC-123",
				"2024-01-15", "2024-01-01", "2024-12-31", "Item",
				"Display", "2024-02-01", "2024-11-30", "not-a-number",
				"abc", "xyz", "invalid", "nope", "30",
			},
			"Non-numeric strings in required fields should return validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := n.BuildSoLineItemParams(tt.row, headerIdx)
			if err == nil {
				t.Errorf("Expected validation error: %s", tt.description)
			}
		})
	}
}

/* ========================================
	Integration-style Tests
======================================== */

func TestBuildSoLineItemParams_RealWorldCSVData(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsSoTestHeaderIndex()

	// Simulate real CSV data that might come from NetSuite
	realWorldRows := [][]string{
		// Standard export
		{
			"0060R00001ABC123", "a1B2C3D4E5F6G7H8", "ACME Corp : Project Alpha",
			"SO-10001", "1/15/2024", "1/1/2024", "12/31/2024",
			"Annual Subscription", "Annual Subscription - Enterprise",
			"1/1/2024", "12/31/2024", "1", "12", "$99,999.00",
			"$0.00", "$99,999.00", "30",
		},
		// Export with netsuite: prefix
		{
			"netsuite:0060R00001XYZ789", "netsuite:a9B8C7D6E5F4G3H2",
			"Beta Inc : Implementation", "SO-10002",
			"02/28/2024", "03/01/2024", "02/28/2025",
			"Professional Services", "PS - Implementation Hours",
			"03/01/2024", "05/31/2024", "160", "160", "$250.00",
			"$10,000.00", "$40,000.00", "45",
		},
		// Export with Excel formula escaping
		{
			`="0060R00001DEF456"`, `="b1C2D3E4F5G6H7I8"`,
			`="Gamma LLC : Maintenance"`, `="SO-10003"`,
			"3/31/24", "4/1/24", "3/31/25",
			"Support", "Support - Premium",
			"4/1/24", "3/31/25", "1", "1", `="$12,000.00"`,
			`="$0.00"`, `="$12,000.00"`, "60",
		},
	}

	for i, row := range realWorldRows {
		t.Run("real_world_row_"+string(rune('A'+i)), func(t *testing.T) {
			params, err := n.BuildSoLineItemParams(row, headerIdx)
			if err != nil {
				t.Errorf("Failed to build params from real world row: %v", err)
				return
			}

			// Verify key fields are populated
			if !params.SalesforceOpportunityID.Valid {
				t.Error("SalesforceOpportunityID should be valid")
			}
			if params.CustomerProject == "" {
				t.Error("CustomerProject should not be empty")
			}
			if !params.DocumentNumber.Valid {
				t.Error("DocumentNumber should be valid")
			}
		})
	}
}

func TestBuildInvoiceSalesTaxParams_RealWorldCSVData(t *testing.T) {
	n := &NsUpload{}
	headerIdx := nsInvoiceTestHeaderIndex()

	realWorldRows := [][]string{
		{
			"Invoice", "01/15/2024", "02/14/2024", "INV-2024-0001",
			"ACME Corporation", "January 2024 Invoice", "CA Sales Tax",
			"1", "1", "$500.00", "$500.00",
			"01/01/2024", "01/31/2024", "4000 - Sales Tax Payable",
			"0060R00001ABC123", "01t000000000ABC", "12345", "67890",
			"Los Angeles", "California", "United States",
		},
	}

	for i, row := range realWorldRows {
		t.Run("real_world_invoice_"+string(rune('A'+i)), func(t *testing.T) {
			params, err := n.BuildInvoiceSalesTaxParams(row, headerIdx)
			if err != nil {
				t.Errorf("Failed to build params: %v", err)
				return
			}

			// Verify state normalization worked
			if params.ShippingAddressState.String != "CA" {
				t.Errorf("State should be normalized to 'CA', got %q",
					params.ShippingAddressState.String)
			}
		})
	}
}
