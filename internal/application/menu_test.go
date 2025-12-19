package application

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

/* ========================================
	linkParents Tests
======================================== */

func TestLinkParents_Positive(t *testing.T) {
	tests := []struct {
		name string
		menu *Menu
	}{
		{
			"single level menu",
			&Menu{
				Title: "Root",
				Items: []MenuItem{
					{Label: "Item 1"},
					{Label: "Item 2"},
				},
			},
		},
		{
			"two level menu",
			&Menu{
				Title: "Root",
				Items: []MenuItem{
					{Label: "Item 1"},
					{Label: "Submenu", Submenu: &Menu{
						Title: "Submenu",
						Items: []MenuItem{
							{Label: "Sub Item 1"},
							{Label: "Back"},
						},
					}},
				},
			},
		},
		{
			"three level menu",
			&Menu{
				Title: "Root",
				Items: []MenuItem{
					{Label: "Level 1", Submenu: &Menu{
						Title: "Level 1",
						Items: []MenuItem{
							{Label: "Level 2", Submenu: &Menu{
								Title: "Level 2",
								Items: []MenuItem{
									{Label: "Deep Item"},
									{Label: "Back"},
								},
							}},
							{Label: "Back"},
						},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkParents(tt.menu, nil)

			// Root should have nil parent
			if tt.menu.Parent != nil {
				t.Error("Root menu should have nil parent")
			}
		})
	}
}

func TestLinkParents_BackButton(t *testing.T) {
	submenu := &Menu{
		Title: "Submenu",
		Items: []MenuItem{
			{Label: "Item"},
			{Label: "Back"},
		},
	}

	root := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Go to submenu", Submenu: submenu},
		},
	}

	linkParents(root, nil)

	// Find the "Back" item in submenu
	var backItem *MenuItem
	for i := range submenu.Items {
		if submenu.Items[i].Label == "Back" {
			backItem = &submenu.Items[i]
			break
		}
	}

	if backItem == nil {
		t.Fatal("Back item not found")
	}

	// Back button should point to root (parent)
	if backItem.Submenu != root {
		t.Error("Back button should point to parent menu")
	}
}

func TestLinkParents_SubmenuParent(t *testing.T) {
	level2 := &Menu{
		Title: "Level 2",
		Items: []MenuItem{
			{Label: "Item"},
		},
	}

	level1 := &Menu{
		Title: "Level 1",
		Items: []MenuItem{
			{Label: "Go to Level 2", Submenu: level2},
		},
	}

	root := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Go to Level 1", Submenu: level1},
		},
	}

	linkParents(root, nil)

	// Verify parent chain
	if root.Parent != nil {
		t.Error("Root should have nil parent")
	}

	if level1.Parent != root {
		t.Error("Level 1 parent should be root")
	}

	if level2.Parent != level1 {
		t.Error("Level 2 parent should be level 1")
	}
}

func TestLinkParents_EmptyMenu(t *testing.T) {
	menu := &Menu{
		Title: "Empty",
		Items: []MenuItem{},
	}

	// Should not panic
	linkParents(menu, nil)

	if menu.Parent != nil {
		t.Error("Empty menu should have nil parent when linked with nil")
	}
}

func TestLinkParents_NilSubmenu(t *testing.T) {
	menu := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Item with nil submenu", Submenu: nil},
			{Label: "Item with action", Action: func() tea.Cmd { return nil }},
		},
	}

	// Should not panic with nil submenus
	linkParents(menu, nil)
}

/* ========================================
	Menu Structure Tests
======================================== */

func TestMenuItem_Label(t *testing.T) {
	item := MenuItem{Label: "Test Label"}

	if item.Label != "Test Label" {
		t.Errorf("Label = %q, want 'Test Label'", item.Label)
	}
}

func TestMenuItem_Submenu(t *testing.T) {
	submenu := &Menu{Title: "Submenu"}
	item := MenuItem{Label: "Has Submenu", Submenu: submenu}

	if item.Submenu == nil {
		t.Error("Submenu should not be nil")
	}

	if item.Submenu.Title != "Submenu" {
		t.Errorf("Submenu.Title = %q, want 'Submenu'", item.Submenu.Title)
	}
}

func TestMenuItem_Action(t *testing.T) {
	actionCalled := false
	item := MenuItem{
		Label: "Has Action",
		Action: func() tea.Cmd {
			return func() tea.Msg {
				actionCalled = true
				return nil
			}
		},
	}

	if item.Action == nil {
		t.Error("Action should not be nil")
	}

	// Call the action
	cmd := item.Action()
	if cmd == nil {
		t.Error("Action should return a command")
	}

	// Execute the command
	cmd()

	if !actionCalled {
		t.Error("Action command should have been executed")
	}
}

/* ========================================
	Edge Cases
======================================== */

func TestLinkParents_CircularPrevention(t *testing.T) {
	// Ensure linkParents doesn't create circular references
	// (it shouldn't, but let's verify)

	menu := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Item"},
		},
	}

	linkParents(menu, nil)
	linkParents(menu, nil) // Call twice

	if menu.Parent != nil {
		t.Error("Calling linkParents twice should still have nil parent")
	}
}

func TestLinkParents_WithExistingParent(t *testing.T) {
	existingParent := &Menu{Title: "Existing Parent"}

	menu := &Menu{
		Title:  "Child",
		Parent: existingParent, // Pre-set parent
		Items:  []MenuItem{{Label: "Item"}},
	}

	// Link with nil should overwrite
	linkParents(menu, nil)

	if menu.Parent != nil {
		t.Error("linkParents should overwrite existing parent with nil")
	}
}

func TestLinkParents_DeepNesting(t *testing.T) {
	// Create deeply nested menu structure
	level5 := &Menu{Title: "Level 5", Items: []MenuItem{{Label: "Deep"}}}
	level4 := &Menu{Title: "Level 4", Items: []MenuItem{{Label: "Sub", Submenu: level5}}}
	level3 := &Menu{Title: "Level 3", Items: []MenuItem{{Label: "Sub", Submenu: level4}}}
	level2 := &Menu{Title: "Level 2", Items: []MenuItem{{Label: "Sub", Submenu: level3}}}
	level1 := &Menu{Title: "Level 1", Items: []MenuItem{{Label: "Sub", Submenu: level2}}}
	root := &Menu{Title: "Root", Items: []MenuItem{{Label: "Sub", Submenu: level1}}}

	linkParents(root, nil)

	// Verify entire chain
	if root.Parent != nil {
		t.Error("Root parent should be nil")
	}
	if level1.Parent != root {
		t.Error("Level 1 parent should be root")
	}
	if level2.Parent != level1 {
		t.Error("Level 2 parent should be level 1")
	}
	if level3.Parent != level2 {
		t.Error("Level 3 parent should be level 2")
	}
	if level4.Parent != level3 {
		t.Error("Level 4 parent should be level 3")
	}
	if level5.Parent != level4 {
		t.Error("Level 5 parent should be level 4")
	}
}

/* ========================================
	False Positive/Negative Tests
======================================== */

func TestLinkParents_MultipleBackItems(t *testing.T) {
	submenu := &Menu{
		Title: "Submenu",
		Items: []MenuItem{
			{Label: "Back"},       // First back
			{Label: "Item"},
			{Label: "Back"},       // Second back
		},
	}

	root := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Sub", Submenu: submenu},
		},
	}

	linkParents(root, nil)

	// Both "Back" items should point to root
	for i, item := range submenu.Items {
		if item.Label == "Back" {
			if item.Submenu != root {
				t.Errorf("Back item %d should point to root", i)
			}
		}
	}
}

func TestLinkParents_BackNotAtEnd(t *testing.T) {
	submenu := &Menu{
		Title: "Submenu",
		Items: []MenuItem{
			{Label: "Back"},       // Back at beginning
			{Label: "Item 1"},
			{Label: "Item 2"},
		},
	}

	root := &Menu{
		Title: "Root",
		Items: []MenuItem{
			{Label: "Sub", Submenu: submenu},
		},
	}

	linkParents(root, nil)

	// Back item should still be linked correctly
	if submenu.Items[0].Submenu != root {
		t.Error("Back item at beginning should still point to parent")
	}
}
