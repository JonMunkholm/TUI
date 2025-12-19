package handler

import (
	"context"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/schema"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NsUpload struct {
	BaseUploader
}

func NewNsUpload(pool *pgxpool.Pool) *NsUpload {
	return &NsUpload{
		BaseUploader: BaseUploader{Pool: pool},
	}
}

func (n *NsUpload) SetProps() error {
	return n.BaseUploader.SetProps("NS", n.makeDirMap)
}

/* ----------------------------------------
	Insert Actions
---------------------------------------- */

func (n *NsUpload) InsertSoLineItem() tea.Cmd {
	return n.RunUpload("SO_line_item_detail")
}

func (n *NsUpload) InsertInvoiceSalesTaxItems() tea.Cmd {
	return n.RunUpload("Invoice_line_item_detail-Sales_Tax")
}

/* ----------------------------------------
	Build Param functions
---------------------------------------- */

func (n *NsUpload) BuildSoLineItemParams(row []string, headerIdx HeaderIndex) (db.InsertNsSoLineItemsParams, error) {
	vrow, err := validateRow(row, headerIdx, schema.NsSoFieldSpecs)
	if err != nil {
		return db.InsertNsSoLineItemsParams{}, err
	}

	return db.InsertNsSoLineItemsParams{
		SalesforceOpportunityID:     ToPgText(vrow["Salesforce Opportunity Id (IO)"]),
		SalesforceOpportunityLineID: ToPgText(vrow["Salesforce Opportunity Line Id (IO)"]),
		CustomerProject:             vrow["Customer/Project"],
		DocumentNumber:              ToPgText(vrow["Document Number"]),
		DocumentDate:                ToPgDate(vrow["Date"]),
		StartDate:                   ToPgDate(vrow["Start Date"]),
		EndDate:                     ToPgDate(vrow["End Date"]),
		ItemName:                    ToPgText(vrow["Item: Name"]),
		ItemDisplayName:             ToPgText(vrow["Item: Display Name"]),
		LineStartDate:               ToPgDate(vrow["Start Date (Line)"]),
		LineEndDate:                 ToPgDate(vrow["End Date (Line Level)"]),
		Quantity:                    ToPgNumeric(vrow["Quantity"]),
		ContractQuantity:            ToPgNumeric(vrow["Contract Quantity"]),
		UnitPrice:                   ToPgNumeric(vrow["Unit Price"]),
		TotalAmountDuePartner:       ToPgNumeric(vrow["Total Amount Due Partner"]),
		AmountGross:                 ToPgNumeric(vrow["Amount (Gross)"]),
		TermsDaysTillNetDue:         ToPgNumeric(vrow["Terms: Days Till Net Due"]),
	}, nil
}

func (n *NsUpload) BuildInvoiceSalesTaxParams(row []string, headerIdx HeaderIndex) (db.InsertNsInvoiceSalesTaxItemsParams, error) {
	vrow, err := validateRow(row, headerIdx, schema.NsInvoiceFieldSpecs)
	if err != nil {
		return db.InsertNsInvoiceSalesTaxItemsParams{}, err
	}

	return db.InsertNsInvoiceSalesTaxItemsParams{
		Type:                    ToPgText(vrow["Type"]),
		Date:                    ToPgDate(vrow["Date"]),
		DateDue:                 ToPgDate(vrow["Date Due"]),
		DocumentNumber:          ToPgText(vrow["Document Number"]),
		Name:                    ToPgText(vrow["Name"]),
		Memo:                    ToPgText(vrow["Memo"]),
		Item:                    ToPgText(vrow["Item"]),
		Qty:                     ToPgNumeric(vrow["Qty"]),
		ContractQuantity:        ToPgNumeric(vrow["Contract Quantity"]),
		UnitPrice:               ToPgNumeric(vrow["Unit Price"]),
		Amount:                  ToPgNumeric(vrow["Amount"]),
		StartDateLine:           ToPgDate(vrow["Start Date (Line)"]),
		EndDateLineLevel:        ToPgDate(vrow["End Date (Line Level)"]),
		Account:                 ToPgText(vrow["Account"]),
		SalesforceOpportunityID: ToPgText(vrow["Salesforce Opportunity Id (IO)"]),
		SalesforcePricebookID:   ToPgText(vrow["Salesforce Pricebook Id (IO)"]),
		ItemInternalID:          ToPgText(vrow["Item: Internal ID"]),
		EntityInternalID:        ToPgText(vrow["Entity: Internal ID"]),
		ShippingAddressCity:     ToPgText(vrow["Address: Shipping Address City"]),
		ShippingAddressState:    ToPgText(vrow["Address: Shipping Address State"]),
		ShippingAddressCountry:  ToPgText(vrow["Address: Shipping Address Country"]),
	}, nil
}

/* ----------------------------------------
	Directory Map
---------------------------------------- */

func (n *NsUpload) makeDirMap() map[string]CsvProps {
	return map[string]CsvProps{
		"Invoice_line_item_detail-Sales_Tax": CsvHandler[db.InsertNsInvoiceSalesTaxItemsParams]{
			specs:  schema.NsInvoiceFieldSpecs,
			build:  n.BuildInvoiceSalesTaxParams,
			insert: n.insertInvoiceSalesTaxItems(),
		},
		"SO_line_item_detail": CsvHandler[db.InsertNsSoLineItemsParams]{
			specs:  schema.NsSoFieldSpecs,
			build:  n.BuildSoLineItemParams,
			insert: n.insertSoLineItems(),
		},
	}
}

/* ----------------------------------------
	Insert Wrappers
---------------------------------------- */

func (n *NsUpload) insertSoLineItems() InsertFn[db.InsertNsSoLineItemsParams] {
	return func(ctx context.Context, queries *db.Queries, arg db.InsertNsSoLineItemsParams) (bool, error) {
		err := queries.InsertNsSoLineItems(ctx, arg)
		return err == nil, err
	}
}

func (n *NsUpload) insertInvoiceSalesTaxItems() InsertFn[db.InsertNsInvoiceSalesTaxItemsParams] {
	return func(ctx context.Context, queries *db.Queries, arg db.InsertNsInvoiceSalesTaxItemsParams) (bool, error) {
		err := queries.InsertNsInvoiceSalesTaxItems(ctx, arg)
		return err == nil, err
	}
}
