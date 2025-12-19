package handler

import (
	"context"
	"fmt"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/schema"
)

/* ----------------------------------------
	HELPER TYPES
---------------------------------------- */

// HeaderIndex maps column names (lowercase) to their position in the CSV row.
// Pre-computed once per file to avoid repeated allocations during row processing.
type HeaderIndex map[string]int

type CsvProps interface {
	Header() []string
	BuildParams(row []string, headerIdx HeaderIndex) (any, error)
	Insert(ctx context.Context, queries *db.Queries, arg any) (bool, error)
}

type BuildParamsFn[T any] func([]string, HeaderIndex) (T, error)
type InsertFn[T any] func(context.Context, *db.Queries, T) (bool, error)

/* ----------------------------------------
	CSV HANDLER WRAPPER
---------------------------------------- */

type CsvHandler[T any] struct {
	specs  []schema.FieldSpec
	build  BuildParamsFn[T]
	insert InsertFn[T]
}

func (h CsvHandler[T]) Header() []string {
	headers := make([]string, len(h.specs))
	for i, s := range h.specs {
		headers[i] = s.Name
	}
	return headers
}

func (h CsvHandler[T]) BuildParams(row []string, headerIdx HeaderIndex) (any, error) {
	return h.build(row, headerIdx)
}

func (h CsvHandler[T]) Insert(ctx context.Context, queries *db.Queries, arg any) (bool, error) {
	typed, ok := arg.(T)
	if !ok {
		return false, fmt.Errorf("invalid param type for handler")
	}

	return h.insert(ctx, queries, typed)
}
