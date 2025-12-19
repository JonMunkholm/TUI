package schema

// NsSoFieldSpecs defines the expected CSV columns for NetSuite SO Line Item reports.
var NsSoFieldSpecs = []FieldSpec{
	{Name: "Salesforce Opportunity Id (IO)", Type: FieldText, Required: true},
	{Name: "Salesforce Opportunity Line Id (IO)", Type: FieldText, Required: true},
	{Name: "Customer/Project", Type: FieldText, Required: true},
	{Name: "Document Number", Type: FieldText, Required: true},
	{Name: "Date", Type: FieldDate, Required: true},
	{Name: "Start Date", Type: FieldDate, Required: true},
	{Name: "End Date", Type: FieldDate, Required: true},
	{Name: "Item: Name", Type: FieldText, Required: true},
	{Name: "Item: Display Name", Type: FieldText, Required: true},
	{Name: "Start Date (Line)", Type: FieldDate, Required: true},
	{Name: "End Date (Line Level)", Type: FieldDate, Required: true},
	{Name: "Quantity", Type: FieldNumeric, Required: true},
	{Name: "Contract Quantity", Type: FieldNumeric, Required: true},
	{Name: "Unit Price", Type: FieldNumeric, Required: true},
	{Name: "Total Amount Due Partner", Type: FieldNumeric, Required: true},
	{Name: "Amount (Gross)", Type: FieldNumeric, Required: true},
	{Name: "Terms: Days Till Net Due", Type: FieldNumeric, Required: true},
}

// NsInvoiceFieldSpecs defines the expected CSV columns for NetSuite Invoice reports.
var NsInvoiceFieldSpecs = []FieldSpec{
	{Name: "Type", Type: FieldText, Required: true},
	{Name: "Date", Type: FieldDate, Required: true},
	{Name: "Date Due", Type: FieldDate, Required: true},
	{Name: "Document Number", Type: FieldText, Required: true},
	{Name: "Name", Type: FieldText, Required: true},
	{Name: "Memo", Type: FieldText, Required: false},
	{Name: "Item", Type: FieldText, Required: true},
	{Name: "Qty", Type: FieldNumeric, Required: true},
	{Name: "Contract Quantity", Type: FieldNumeric, Required: true},
	{Name: "Unit Price", Type: FieldNumeric, Required: true},
	{Name: "Amount", Type: FieldNumeric, Required: true},
	{Name: "Start Date (Line)", Type: FieldDate, Required: true},
	{Name: "End Date (Line Level)", Type: FieldDate, Required: true},
	{Name: "Account", Type: FieldText, Required: true},
	{Name: "Salesforce Opportunity Id (IO)", Type: FieldText, Required: true},
	{Name: "Salesforce Pricebook Id (IO)", Type: FieldText, Required: true},
	{Name: "Item: Internal ID", Type: FieldText, Required: true},
	{Name: "Entity: Internal ID", Type: FieldText, Required: true},
	{Name: "Address: Shipping Address City", Type: FieldText, Required: true},
	{Name: "Address: Shipping Address State", Type: FieldText, Required: true, Normalizer: NormalizeUsState},
	{Name: "Address: Shipping Address Country", Type: FieldText, Required: true},
}
