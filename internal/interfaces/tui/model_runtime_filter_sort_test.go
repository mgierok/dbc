package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_ShiftSOpensSortPopupInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		tables:   []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if !model.sortPopup.active {
		t.Fatal("expected sort popup to be active")
	}
	if model.sortPopup.step != sortSelectColumn {
		t.Fatalf("expected sort popup to start at column step, got %v", model.sortPopup.step)
	}
}

func TestHandleKey_ShiftSIgnoredOutsideRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewSchema,
		focus:    FocusContent,
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if model.sortPopup.active {
		t.Fatal("expected sort popup to stay closed outside records view")
	}
}

func TestHandleKey_ShiftFOpensFilterPopupInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode: ViewRecords,
		focus:    FocusContent,
		tables:   []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

	// Assert
	if !model.filterPopup.active {
		t.Fatal("expected filter popup to be active")
	}
	if model.filterPopup.step != filterSelectColumn {
		t.Fatalf("expected filter popup to start at column step, got %v", model.filterPopup.step)
	}
}

func TestHandleKey_ShiftFIgnoredOutsideRecordsContext(t *testing.T) {
	tests := []struct {
		name     string
		viewMode ViewMode
		focus    PanelFocus
	}{
		{
			name:     "schema content",
			viewMode: ViewSchema,
			focus:    FocusContent,
		},
		{
			name:     "records tables panel",
			viewMode: ViewRecords,
			focus:    FocusTables,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				viewMode: tc.viewMode,
				focus:    tc.focus,
				tables:   []dto.Table{{Name: "users"}},
				schema: dto.Schema{
					Columns: []dto.SchemaColumn{
						{Name: "id", Type: "INTEGER"},
						{Name: "name", Type: "TEXT"},
					},
				},
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

			// Assert
			if model.filterPopup.active {
				t.Fatal("expected filter popup to stay closed outside records context")
			}
		})
	}
}

func TestHandleFilterPopupKey_EnterProgressesStepsAndAppliesFilter(t *testing.T) {
	// Arrange
	operatorsSpy := &spyListOperatorsUseCase{
		operators: []dto.Operator{
			{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		},
	}
	recordsSpy := &spyListRecordsUseCase{}
	model := &Model{
		ctx:           context.Background(),
		viewMode:      ViewRecords,
		focus:         FocusContent,
		listOperators: operatorsSpy,
		listRecords:   recordsSpy,
		tables:        []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
		filterPopup: filterPopup{
			active:      true,
			step:        filterSelectColumn,
			columnIndex: 1,
		},
		recordPageIndex: 3,
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving from column to operator step")
	}
	if model.filterPopup.step != filterSelectOperator {
		t.Fatalf("expected operator step, got %v", model.filterPopup.step)
	}
	if operatorsSpy.lastColumnType != "TEXT" {
		t.Fatalf("expected operator lookup for TEXT column, got %q", operatorsSpy.lastColumnType)
	}

	// Act
	_, cmd = model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving to value-input step")
	}
	if model.filterPopup.step != filterInputValue {
		t.Fatalf("expected value-input step, got %v", model.filterPopup.step)
	}

	// Act
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alice")})
	_, cmd = model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after filter apply")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.filterPopup.active {
		t.Fatal("expected filter popup to close after apply")
	}
	if model.currentFilter == nil {
		t.Fatal("expected current filter to be set")
	}
	if model.currentFilter.Column != "name" {
		t.Fatalf("expected filter column name, got %q", model.currentFilter.Column)
	}
	if model.currentFilter.Value != "alice" {
		t.Fatalf("expected filter value alice, got %q", model.currentFilter.Value)
	}
	if model.currentFilter.Operator.Kind != dto.OperatorKindEq {
		t.Fatalf("expected operator kind %q, got %q", dto.OperatorKindEq, model.currentFilter.Operator.Kind)
	}
	if model.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after filter apply, got %d", model.recordPageIndex)
	}
	if recordsSpy.lastFilter == nil {
		t.Fatal("expected filter forwarded to list-records use case")
	}
	if recordsSpy.lastFilter.Column != "name" {
		t.Fatalf("expected forwarded filter column name, got %q", recordsSpy.lastFilter.Column)
	}
	if recordsSpy.lastFilter.Value != "alice" {
		t.Fatalf("expected forwarded filter value alice, got %q", recordsSpy.lastFilter.Value)
	}
}

func TestHandleFilterPopupKey_InputEditingSupportsCursorMovementAndBackspace(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "ac",
			cursor: 1,
		},
	}

	// Act
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyLeft})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyBackspace})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyLeft})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})
	model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyRight})

	// Assert
	if model.filterPopup.input != "bc" {
		t.Fatalf("expected edited input bc, got %q", model.filterPopup.input)
	}
	if model.filterPopup.cursor != 2 {
		t.Fatalf("expected cursor clamped at input end, got %d", model.filterPopup.cursor)
	}
}

func TestHandleFilterPopupKey_EscClosesPopup(t *testing.T) {
	// Arrange
	model := &Model{
		filterPopup: filterPopup{
			active: true,
			step:   filterSelectOperator,
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when closing filter popup")
	}
	if model.filterPopup.active {
		t.Fatal("expected filter popup to close on Esc")
	}
}

func TestHandleKey_ShiftSApplySortReloadsRecords(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
		},
	}
	model := &Model{
		ctx:         context.Background(),
		viewMode:    ViewRecords,
		focus:       FocusContent,
		listRecords: recordsSpy,
		tables:      []dto.Table{{Name: "users"}},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "name", Type: "TEXT"},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after applying sort")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.sortPopup.active {
		t.Fatal("expected sort popup to close after apply")
	}
	if model.currentSort == nil {
		t.Fatal("expected current sort to be set")
	}
	if model.currentSort.Column != "id" {
		t.Fatalf("expected sorted column id, got %q", model.currentSort.Column)
	}
	if model.currentSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected sort direction DESC, got %s", model.currentSort.Direction)
	}
	if recordsSpy.lastSort == nil {
		t.Fatal("expected engine to receive sort")
	}
	if recordsSpy.lastSort.Column != "id" || recordsSpy.lastSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected engine sort id DESC, got %+v", recordsSpy.lastSort)
	}
}
