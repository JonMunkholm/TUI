// Package admin provides administrative operations for database management.
package admin

import (
	"context"
	"time"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/handler"
	tea "github.com/charmbracelet/bubbletea"
)

// ResetTimeout is the maximum duration for database reset operations.
const ResetTimeout = 30 * time.Second

// ResetDbs handles database reset operations.
type ResetDbs struct {
	DB *db.Queries
}

type dbResetFn func(ctx context.Context) error

// ResetAll truncates all data tables and clears the CSV upload log.
// This is a destructive operation - use with caution.
func (r *ResetDbs) ResetAll() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), ResetTimeout)
		defer cancel()

		if err := r.runResets(ctx, []dbResetFn{
			r.DB.ResetNsSoLineItems,
			r.DB.ResetNsInvoiceSalesTaxItems,
			r.DB.ResetSfdcOppLineItems,
			r.DB.ResetAnrokTransactions,
			r.DB.ResetCsvUpload,
		}); err != nil {
			return handler.ErrMsg{Err: err}
		}

		return handler.DoneMsg("DBs reset")
	}
}

func (r *ResetDbs) runResets(ctx context.Context, resets []dbResetFn) error {
	for _, reset := range resets {
		if err := reset(ctx); err != nil {
			return err
		}
	}
	return nil
}
