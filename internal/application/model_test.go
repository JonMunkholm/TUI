package application

import (
	"errors"
	"testing"

	"github.com/JonMunkholm/TUI/internal/handler"
	tea "github.com/charmbracelet/bubbletea"
)

/* ========================================
	handleResultMessage Tests
======================================== */

func TestHandleResultMessage_DoneMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	msg := handler.DoneMsg("Operation completed!")

	result := handleResultMessage(m, msg)

	if !result {
		t.Error("handleResultMessage should return true for DoneMsg")
	}

	if m.loading {
		t.Error("loading should be false after DoneMsg")
	}

	if m.output != "Operation completed!" {
		t.Errorf("output = %q, want 'Operation completed!'", m.output)
	}
}

func TestHandleResultMessage_WdMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	msg := handler.WdMsg("Working directory message")

	result := handleResultMessage(m, msg)

	if !result {
		t.Error("handleResultMessage should return true for WdMsg")
	}

	if m.loading {
		t.Error("loading should be false after WdMsg")
	}

	if m.output != "Working directory message" {
		t.Errorf("output = %q, want 'Working directory message'", m.output)
	}
}

func TestHandleResultMessage_ErrMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	testErr := errors.New("something went wrong")
	msg := handler.ErrMsg{Err: testErr}

	result := handleResultMessage(m, msg)

	if !result {
		t.Error("handleResultMessage should return true for ErrMsg")
	}

	if m.loading {
		t.Error("loading should be false after ErrMsg")
	}

	expectedOutput := "Error: something went wrong"
	if m.output != expectedOutput {
		t.Errorf("output = %q, want %q", m.output, expectedOutput)
	}
}

func TestHandleResultMessage_UnknownMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "initial",
	}

	// Unknown message type
	msg := "just a string"

	result := handleResultMessage(m, msg)

	if result {
		t.Error("handleResultMessage should return false for unknown message type")
	}

	// State should be unchanged
	if !m.loading {
		t.Error("loading should remain true for unknown message")
	}

	if m.output != "initial" {
		t.Error("output should remain unchanged for unknown message")
	}
}

func TestHandleResultMessage_NilMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	result := handleResultMessage(m, nil)

	if result {
		t.Error("handleResultMessage should return false for nil message")
	}
}

/* ========================================
	HandleTeaCmdErrorWithTitle Tests
======================================== */

func TestHandleTeaCmdErrorWithTitle_Success(t *testing.T) {
	parent := &Menu{Title: "Parent"}
	title := "Test Title"

	successCmd := func() tea.Msg {
		return handler.DoneMsg("Success!")
	}

	wrappedCmd := HandleTeaCmdErrorWithTitle(title, parent, successCmd)

	// Execute the wrapped command
	msg := wrappedCmd()

	// Should return the original success message
	if _, ok := msg.(handler.DoneMsg); !ok {
		t.Error("Should return DoneMsg for successful command")
	}
}

func TestHandleTeaCmdErrorWithTitle_Error(t *testing.T) {
	parent := &Menu{Title: "Parent"}
	title := "Error Title"

	testErr := errors.New("test error")
	errorCmd := func() tea.Msg {
		return handler.ErrMsg{Err: testErr}
	}

	wrappedCmd := HandleTeaCmdErrorWithTitle(title, parent, errorCmd)

	// Execute the wrapped command
	msg := wrappedCmd()

	// Should return MenuMsg with error menu
	menuMsg, ok := msg.(MenuMsg)
	if !ok {
		t.Fatal("Should return MenuMsg for error")
	}

	// Verify error menu structure
	if menuMsg.Menu.Title != title {
		t.Errorf("Menu.Title = %q, want %q", menuMsg.Menu.Title, title)
	}

	if menuMsg.Menu.Parent != parent {
		t.Error("Menu.Parent should be the parent passed in")
	}

	// Should have error message and Back items
	if len(menuMsg.Menu.Items) != 2 {
		t.Errorf("Menu should have 2 items, got %d", len(menuMsg.Menu.Items))
	}

	// First item should be error message
	if menuMsg.Menu.Items[0].Label != "Error: test error" {
		t.Errorf("First item = %q, want 'Error: test error'", menuMsg.Menu.Items[0].Label)
	}

	// Second item should be Back
	if menuMsg.Menu.Items[1].Label != "Back" {
		t.Errorf("Second item = %q, want 'Back'", menuMsg.Menu.Items[1].Label)
	}
}

func TestHandleTeaCmdErrorWithTitle_NilParent(t *testing.T) {
	title := "Test"

	errorCmd := func() tea.Msg {
		return handler.ErrMsg{Err: errors.New("error")}
	}

	wrappedCmd := HandleTeaCmdErrorWithTitle(title, nil, errorCmd)

	// Should not panic with nil parent
	msg := wrappedCmd()

	menuMsg, ok := msg.(MenuMsg)
	if !ok {
		t.Fatal("Should return MenuMsg")
	}

	if menuMsg.Menu.Parent != nil {
		t.Error("Menu.Parent should be nil when passed nil")
	}
}

/* ========================================
	MenuMsg Tests
======================================== */

func TestMenuMsg_Structure(t *testing.T) {
	menu := &Menu{
		Title: "Test Menu",
		Items: []MenuItem{
			{Label: "Item 1"},
		},
	}

	msg := MenuMsg{Menu: menu}

	if msg.Menu != menu {
		t.Error("MenuMsg.Menu should reference the passed menu")
	}

	if msg.Menu.Title != "Test Menu" {
		t.Errorf("Menu.Title = %q, want 'Test Menu'", msg.Menu.Title)
	}
}

/* ========================================
	Edge Cases
======================================== */

func TestHandleResultMessage_EmptyError(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	msg := handler.ErrMsg{Err: errors.New("")}

	result := handleResultMessage(m, msg)

	if !result {
		t.Error("Should handle empty error message")
	}

	if m.output != "Error: " {
		t.Errorf("output = %q, want 'Error: '", m.output)
	}
}

func TestHandleResultMessage_EmptyDoneMsg(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	msg := handler.DoneMsg("")

	result := handleResultMessage(m, msg)

	if !result {
		t.Error("Should handle empty DoneMsg")
	}

	if m.output != "" {
		t.Errorf("output = %q, want ''", m.output)
	}
}

/* ========================================
	False Positive/Negative Tests
======================================== */

func TestHandleResultMessage_TypedNil(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "initial",
	}

	// Typed nil - this could be tricky
	var errMsg *handler.ErrMsg = nil

	result := handleResultMessage(m, errMsg)

	// Should not match ErrMsg case because it's nil
	if result {
		t.Error("Should return false for typed nil")
	}
}

func TestHandleTeaCmdErrorWithTitle_NonErrorMsg(t *testing.T) {
	parent := &Menu{Title: "Parent"}

	// Test with WdMsg (not an error)
	wdCmd := func() tea.Msg {
		return handler.WdMsg("Working dir")
	}

	wrappedCmd := HandleTeaCmdErrorWithTitle("Title", parent, wdCmd)
	msg := wrappedCmd()

	// Should pass through WdMsg unchanged
	if _, ok := msg.(handler.WdMsg); !ok {
		t.Error("WdMsg should pass through unchanged")
	}
}

func TestHandleTeaCmdErrorWithTitle_MenuMsgPassthrough(t *testing.T) {
	parent := &Menu{Title: "Parent"}
	otherMenu := &Menu{Title: "Other Menu"}

	// Test with MenuMsg (should pass through)
	menuCmd := func() tea.Msg {
		return MenuMsg{Menu: otherMenu}
	}

	wrappedCmd := HandleTeaCmdErrorWithTitle("Title", parent, menuCmd)
	msg := wrappedCmd()

	// Should pass through MenuMsg
	menuMsg, ok := msg.(MenuMsg)
	if !ok {
		t.Fatal("MenuMsg should pass through")
	}

	if menuMsg.Menu != otherMenu {
		t.Error("Original MenuMsg should pass through unchanged")
	}
}

/* ========================================
	Integration Tests
======================================== */

func TestHandleResultMessage_StateTransitions(t *testing.T) {
	m := &Model{
		loading: true,
		output:  "",
	}

	// First: process DoneMsg
	handleResultMessage(m, handler.DoneMsg("First"))
	if m.loading || m.output != "First" {
		t.Error("State incorrect after first message")
	}

	// Reset state
	m.loading = true

	// Second: process ErrMsg
	handleResultMessage(m, handler.ErrMsg{Err: errors.New("second")})
	if m.loading || m.output != "Error: second" {
		t.Error("State incorrect after second message")
	}

	// Reset state
	m.loading = true

	// Third: process WdMsg
	handleResultMessage(m, handler.WdMsg("third"))
	if m.loading || m.output != "third" {
		t.Error("State incorrect after third message")
	}
}
