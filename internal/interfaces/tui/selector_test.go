package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDatabaseSelector_EnterSelects(t *testing.T) {
	// Arrange
	model := newDatabaseSelectorModel([]DatabaseOption{
		{Name: "local", ConnString: "/tmp/example.sqlite"},
	})

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if !selector.chosen {
		t.Fatal("expected selection to be confirmed")
	}
}

func TestDatabaseSelector_EscCancels(t *testing.T) {
	// Arrange
	model := newDatabaseSelectorModel([]DatabaseOption{
		{Name: "local", ConnString: "/tmp/example.sqlite"},
	})

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if !selector.canceled {
		t.Fatal("expected selection to be canceled")
	}
}

func TestDatabaseSelector_MoveSelection(t *testing.T) {
	// Arrange
	model := newDatabaseSelectorModel([]DatabaseOption{
		{Name: "local", ConnString: "/tmp/example.sqlite"},
		{Name: "analytics", ConnString: "/tmp/analytics.sqlite"},
	})

	// Act
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Assert
	selector := updated.(*databaseSelectorModel)
	if selector.selected != 1 {
		t.Fatalf("expected selection to move down, got %d", selector.selected)
	}

	// Act
	updated, _ = selector.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Assert
	selector = updated.(*databaseSelectorModel)
	if selector.selected != 0 {
		t.Fatalf("expected selection to move up, got %d", selector.selected)
	}
}
