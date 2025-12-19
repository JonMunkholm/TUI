package handler

import (
	"context"
	"errors"
	"testing"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/schema"
)

/* ========================================
	CsvHandler Tests
======================================== */

// Test struct for CsvHandler tests
type testParams struct {
	Field1 string
	Field2 int
}

// Helper to create FieldSpecs from header names
func specsFromHeaders(headers []string) []schema.FieldSpec {
	specs := make([]schema.FieldSpec, len(headers))
	for i, h := range headers {
		specs[i] = schema.FieldSpec{Name: h, Type: schema.FieldText, Required: false}
	}
	return specs
}

func TestCsvHandler_Header(t *testing.T) {
	tests := []struct {
		name     string
		specs    []schema.FieldSpec
		expected []string
	}{
		{
			"single header",
			specsFromHeaders([]string{"Column1"}),
			[]string{"Column1"},
		},
		{
			"multiple headers",
			specsFromHeaders([]string{"Column1", "Column2", "Column3"}),
			[]string{"Column1", "Column2", "Column3"},
		},
		{
			"empty header",
			specsFromHeaders([]string{}),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CsvHandler[testParams]{
				specs: tt.specs,
			}

			result := handler.Header()

			if len(result) != len(tt.expected) {
				t.Errorf("Header() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Header()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestCsvHandler_BuildParams_Positive(t *testing.T) {
	buildFn := func(row []string, _ HeaderIndex) (testParams, error) {
		return testParams{
			Field1: row[0],
			Field2: len(row),
		}, nil
	}

	handler := CsvHandler[testParams]{
		specs: specsFromHeaders([]string{"Column1", "Column2"}),
		build: buildFn,
	}

	row := []string{"value1", "value2"}
	headerIdx := MakeHeaderIndex(handler.Header())
	result, err := handler.BuildParams(row, headerIdx)

	if err != nil {
		t.Errorf("BuildParams() error = %v", err)
		return
	}

	params, ok := result.(testParams)
	if !ok {
		t.Errorf("BuildParams() returned wrong type")
		return
	}

	if params.Field1 != "value1" {
		t.Errorf("Field1 = %q, want 'value1'", params.Field1)
	}

	if params.Field2 != 2 {
		t.Errorf("Field2 = %d, want 2", params.Field2)
	}
}

func TestCsvHandler_BuildParams_Negative(t *testing.T) {
	expectedErr := errors.New("build error")
	buildFn := func(row []string, _ HeaderIndex) (testParams, error) {
		return testParams{}, expectedErr
	}

	handler := CsvHandler[testParams]{
		build: buildFn,
	}

	headerIdx := MakeHeaderIndex(handler.Header())
	_, err := handler.BuildParams([]string{"value"}, headerIdx)

	if err == nil {
		t.Error("BuildParams() expected error, got nil")
		return
	}

	if err != expectedErr {
		t.Errorf("BuildParams() error = %v, want %v", err, expectedErr)
	}
}

func TestCsvHandler_Insert_Positive(t *testing.T) {
	insertCalled := false
	insertFn := func(ctx context.Context, queries *db.Queries, arg testParams) (bool, error) {
		insertCalled = true
		return true, nil
	}

	handler := CsvHandler[testParams]{
		insert: insertFn,
	}

	ctx := context.Background()
	params := testParams{Field1: "test", Field2: 42}

	success, err := handler.Insert(ctx, nil, params) // nil queries for unit test

	if err != nil {
		t.Errorf("Insert() error = %v", err)
	}

	if !success {
		t.Error("Insert() returned false, want true")
	}

	if !insertCalled {
		t.Error("Insert function was not called")
	}
}

func TestCsvHandler_Insert_Negative_WrongType(t *testing.T) {
	insertFn := func(ctx context.Context, queries *db.Queries, arg testParams) (bool, error) {
		return true, nil
	}

	handler := CsvHandler[testParams]{
		insert: insertFn,
	}

	ctx := context.Background()
	wrongType := "not a testParams"

	success, err := handler.Insert(ctx, nil, wrongType)

	if err == nil {
		t.Error("Insert() expected error for wrong type, got nil")
		return
	}

	if success {
		t.Error("Insert() should return false on type error")
	}

	expectedMsg := "invalid param type for handler"
	if err.Error() != expectedMsg {
		t.Errorf("Insert() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestCsvHandler_Insert_Negative_InsertError(t *testing.T) {
	expectedErr := errors.New("database error")
	insertFn := func(ctx context.Context, queries *db.Queries, arg testParams) (bool, error) {
		return false, expectedErr
	}

	handler := CsvHandler[testParams]{
		insert: insertFn,
	}

	ctx := context.Background()
	params := testParams{Field1: "test", Field2: 42}

	success, err := handler.Insert(ctx, nil, params)

	if err != expectedErr {
		t.Errorf("Insert() error = %v, want %v", err, expectedErr)
	}

	if success {
		t.Error("Insert() should return false on error")
	}
}

/* ========================================
	CsvProps Interface Tests
======================================== */

func TestCsvHandler_ImplementsCsvProps(t *testing.T) {
	// Verify CsvHandler implements CsvProps interface
	var _ CsvProps = CsvHandler[testParams]{}
}

/* ========================================
	Integration Tests
======================================== */

func TestCsvHandler_FullWorkflow(t *testing.T) {
	// Simulate a full CSV processing workflow

	type mockDbParams struct {
		Name  string
		Value float64
	}

	buildCount := 0
	insertCount := 0

	handler := CsvHandler[mockDbParams]{
		specs: specsFromHeaders([]string{"Name", "Value"}),
		build: func(row []string, _ HeaderIndex) (mockDbParams, error) {
			buildCount++
			if len(row) < 2 {
				return mockDbParams{}, errors.New("row too short")
			}
			return mockDbParams{
				Name:  row[0],
				Value: 100.0,
			}, nil
		},
		insert: func(ctx context.Context, queries *db.Queries, arg mockDbParams) (bool, error) {
			insertCount++
			if arg.Name == "" {
				return false, errors.New("empty name")
			}
			return true, nil
		},
	}

	// Test header
	if len(handler.Header()) != 2 {
		t.Error("Header should have 2 elements")
	}

	// Test build and insert workflow
	testRows := [][]string{
		{"Product A", "100.00"},
		{"Product B", "200.00"},
		{"Product C", "300.00"},
	}

	ctx := context.Background()
	headerIdx := MakeHeaderIndex(handler.Header()) // Pre-compute once

	for _, row := range testRows {
		arg, err := handler.BuildParams(row, headerIdx)
		if err != nil {
			t.Errorf("BuildParams failed: %v", err)
			continue
		}

		success, err := handler.Insert(ctx, nil, arg)
		if err != nil {
			t.Errorf("Insert failed: %v", err)
			continue
		}

		if !success {
			t.Error("Insert should succeed")
		}
	}

	if buildCount != 3 {
		t.Errorf("Build called %d times, want 3", buildCount)
	}

	if insertCount != 3 {
		t.Errorf("Insert called %d times, want 3", insertCount)
	}
}

/* ========================================
	Edge Cases
======================================== */

func TestCsvHandler_NilFunctions(t *testing.T) {
	// Test behavior when build/insert functions are nil
	handler := CsvHandler[testParams]{
		specs:  specsFromHeaders([]string{"Column1"}),
		build:  nil,
		insert: nil,
	}

	// Header should still work
	if len(handler.Header()) != 1 {
		t.Error("Header should work even with nil functions")
	}

	// BuildParams with nil build function will panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("BuildParams with nil function should panic")
		}
	}()

	headerIdx := MakeHeaderIndex(handler.Header())
	_, _ = handler.BuildParams([]string{"test"}, headerIdx)
}

func TestCsvHandler_EmptyRow(t *testing.T) {
	handler := CsvHandler[testParams]{
		build: func(row []string, _ HeaderIndex) (testParams, error) {
			if len(row) == 0 {
				return testParams{}, errors.New("empty row")
			}
			return testParams{Field1: row[0]}, nil
		},
	}

	headerIdx := MakeHeaderIndex(handler.Header())
	_, err := handler.BuildParams([]string{}, headerIdx)

	if err == nil {
		t.Error("BuildParams should return error for empty row")
	}
}

/* ========================================
	Type Safety Tests
======================================== */

func TestCsvHandler_TypeSafety(t *testing.T) {
	// Test that the generic type constraint works correctly

	type params1 struct{ A string }
	type params2 struct{ B int }

	handler1 := CsvHandler[params1]{
		insert: func(ctx context.Context, queries *db.Queries, arg params1) (bool, error) {
			return arg.A == "test", nil
		},
	}

	handler2 := CsvHandler[params2]{
		insert: func(ctx context.Context, queries *db.Queries, arg params2) (bool, error) {
			return arg.B == 42, nil
		},
	}

	ctx := context.Background()

	// Correct types should work
	success1, _ := handler1.Insert(ctx, nil, params1{A: "test"})
	if !success1 {
		t.Error("handler1 should accept params1")
	}

	success2, _ := handler2.Insert(ctx, nil, params2{B: 42})
	if !success2 {
		t.Error("handler2 should accept params2")
	}

	// Wrong types should fail
	_, err1 := handler1.Insert(ctx, nil, params2{B: 42})
	if err1 == nil {
		t.Error("handler1 should reject params2")
	}

	_, err2 := handler2.Insert(ctx, nil, params1{A: "test"})
	if err2 == nil {
		t.Error("handler2 should reject params1")
	}
}
