package handler

import (
	"context"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/schema"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SfdcUpload struct {
	BaseUploader
}

func NewSfdcUpload(pool *pgxpool.Pool) *SfdcUpload {
	return &SfdcUpload{
		BaseUploader: BaseUploader{Pool: pool},
	}
}

func (s *SfdcUpload) SetProps() error {
	return s.BaseUploader.SetProps("SFDC", s.makeDirMap)
}

/* ----------------------------------------
	Insert Actions
---------------------------------------- */

func (s *SfdcUpload) InsertClosedWonProds() tea.Cmd {
	return s.RunUpload("Closed_Won_Ops-Products_Report")
}

/* ----------------------------------------
	Build Param functions
---------------------------------------- */

func (s *SfdcUpload) BuildSfdcOppLineItemParams(row []string, headerIdx HeaderIndex) (db.InsertSfdcOppLineItemsParams, error) {
	vrow, err := validateRow(row, headerIdx, schema.SfdcFieldSpecs)
	if err != nil {
		return db.InsertSfdcOppLineItemsParams{}, err
	}

	return db.InsertSfdcOppLineItemsParams{
		OpportunityID:                ToPgText(vrow["Opportunity ID Casesafe"]),
		OpportunityProductCasesafeID: ToPgText(vrow["Opportunity Product Casesafe ID"]),
		OpportunityName:              ToPgText(vrow["Opportunity Name"]),
		AccountName:                  ToPgText(vrow["Account Name"]),
		CloseDate:                    ToPgDate(vrow["Close Date"]),
		BookedDate:                   ToPgDate(vrow["Booked Date"]),
		FiscalPeriod:                 ToPgText(vrow["Fiscal Period"]),
		PaymentSchedule:              ToPgText(vrow["Payment Schedule"]),
		PaymentDue:                   ToPgText(vrow["Payment Due"]),
		ContractStartDate:            ToPgDate(vrow["Contract Start Date"]),
		ContractEndDate:              ToPgDate(vrow["Contract End Date"]),
		TermInMonthsDeprecated:       ToPgNumeric(vrow["Term in Months_deprecated"]),
		ProductName:                  ToPgText(vrow["Product Name"]),
		DeploymentType:               ToPgText(vrow["Deployment Type"]),
		Amount:                       ToPgNumeric(vrow["Amount"]),
		Quantity:                     ToPgNumeric(vrow["Quantity"]),
		ListPrice:                    ToPgNumeric(vrow["List Price"]),
		SalesPrice:                   ToPgNumeric(vrow["Sales Price"]),
		TotalPrice:                   ToPgNumeric(vrow["Total Price"]),
		StartDate:                    ToPgDate(vrow["Start Date"]),
		EndDate:                      ToPgDate(vrow["End Date"]),
		TermInMonths:                 ToPgNumeric(vrow["Term in Months"]),
		ProductCode:                  ToPgText(vrow["Product Code"]),
		TotalAmountDueCustomer:       ToPgNumeric(vrow["Total Amount Due - Customer"]),
		TotalAmountDuePartner:        ToPgNumeric(vrow["Total Amount Due - Partner"]),
		ActiveProduct:                ToPgBool(vrow["Active Product"]),
	}, nil
}

/* ----------------------------------------
	Directory Map
---------------------------------------- */

func (s *SfdcUpload) makeDirMap() map[string]CsvProps {
	return map[string]CsvProps{
		"Closed_Won_Ops-Products_Report": CsvHandler[db.InsertSfdcOppLineItemsParams]{
			specs:  schema.SfdcFieldSpecs,
			build:  s.BuildSfdcOppLineItemParams,
			insert: s.insertSfdcOppLineItems(),
		},
	}
}

/* ----------------------------------------
	Insert Wrapper
---------------------------------------- */

func (s *SfdcUpload) insertSfdcOppLineItems() InsertFn[db.InsertSfdcOppLineItemsParams] {
	return func(ctx context.Context, queries *db.Queries, arg db.InsertSfdcOppLineItemsParams) (bool, error) {
		err := queries.InsertSfdcOppLineItems(ctx, arg)
		return err == nil, err
	}
}
