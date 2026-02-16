package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestStatusShortcuts_TablesPanel(t *testing.T) {
	// Arrange
	model := &Model{focus: FocusTables}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Tables: F filter" {
		t.Fatalf("expected table shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_Popup(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{active: true},
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Popup: Enter apply | Esc close" {
		t.Fatalf("expected popup shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_RecordsPanel(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewRecords,
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Records: Enter edit | i insert | d delete | u undo | Ctrl+r redo | w save | F filter" {
		t.Fatalf("expected records shortcuts, got %q", shortcuts)
	}
}

func TestRenderStatus_ShowsDirtyCount(t *testing.T) {
	// Arrange
	model := &Model{
		pendingUpdates: map[string]recordEdits{
			"id=1": {
				changes: map[int]stagedEdit{
					0: {Value: domainmodel.Value{Text: "bob", Raw: "bob"}},
				},
			},
		},
		pendingInserts: []pendingInsertRow{{}},
		pendingDeletes: map[string]recordDelete{"id=2": {}},
	}

	// Act
	status := model.renderStatus(80)

	// Assert
	if !strings.Contains(status, "WRITE (dirty: 3)") {
		t.Fatalf("expected dirty status, got %q", status)
	}
}

func TestRenderStatus_CommandPromptShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		commandInput: commandInput{
			active: true,
			value:  "config",
			cursor: 3,
		},
	}

	// Act
	status := model.renderStatus(200)

	// Assert
	if !strings.Contains(status, "Command: :con|fig") {
		t.Fatalf("expected command prompt caret in status, got %q", status)
	}
}

func TestRenderFilterPopup_ValueInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "abc",
			cursor: 1,
		},
	}

	// Act
	popup := strings.Join(model.renderFilterPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Value: a|bc") {
		t.Fatalf("expected caret in filter value input, got %q", popup)
	}
}

func TestRenderEditPopup_TextInputShowsCaretAtCursor(t *testing.T) {
	// Arrange
	model := &Model{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{
					Name:  "name",
					Type:  "TEXT",
					Input: dto.ColumnInput{Kind: dto.ColumnInputText},
				},
			},
		},
		editPopup: editPopup{
			active:      true,
			columnIndex: 0,
			input:       "john",
			cursor:      2,
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Value: jo|hn") {
		t.Fatalf("expected caret in edit text input, got %q", popup)
	}
}
