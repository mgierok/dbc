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
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if !model.overlay.sortPopup.active {
		t.Fatal("expected sort popup to be active")
	}
	if model.overlay.sortPopup.step != sortSelectColumn {
		t.Fatalf("expected sort popup to start at column step, got %v", model.overlay.sortPopup.step)
	}
}

func TestHandleKey_ShiftSIgnoredOutsideRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusContent,
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})

	// Assert
	if model.overlay.sortPopup.active {
		t.Fatal("expected sort popup to stay closed outside records view")
	}
}

func TestHandleKey_ShiftFOpensFilterPopupInRecordsView(t *testing.T) {
	// Arrange
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

	// Assert
	if !model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to be active")
	}
	if model.overlay.filterPopup.step != filterSelectColumn {
		t.Fatalf("expected filter popup to start at column step, got %v", model.overlay.filterPopup.step)
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
				read: runtimeReadState{
					viewMode: tc.viewMode,
					focus:    tc.focus,
					tables:   []dto.Table{{Name: "users"}},
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER"},
							{Name: "name", Type: "TEXT"},
						},
					},
				},
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})

			// Assert
			if model.overlay.filterPopup.active {
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
		listOperators: operatorsSpy,
		listRecords:   recordsSpy,
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
			},
			recordPageIndex: 3,
		},
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active:      true,
				step:        filterSelectColumn,
				columnIndex: 1,
			},
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving from column to operator step")
	}
	if model.overlay.filterPopup.step != filterSelectOperator {
		t.Fatalf("expected operator step, got %v", model.overlay.filterPopup.step)
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
	if model.overlay.filterPopup.step != filterInputValue {
		t.Fatalf("expected value-input step, got %v", model.overlay.filterPopup.step)
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
	if model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to close after apply")
	}
	if model.read.currentFilter == nil {
		t.Fatal("expected current filter to be set")
	}
	if model.read.currentFilter.Column != "name" {
		t.Fatalf("expected filter column name, got %q", model.read.currentFilter.Column)
	}
	if model.read.currentFilter.Value != "alice" {
		t.Fatalf("expected filter value alice, got %q", model.read.currentFilter.Value)
	}
	if model.read.currentFilter.Operator.Kind != dto.OperatorKindEq {
		t.Fatalf("expected operator kind %q, got %q", dto.OperatorKindEq, model.read.currentFilter.Operator.Kind)
	}
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after filter apply, got %d", model.read.recordPageIndex)
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
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active: true,
				step:   filterInputValue,
				input:  "ac",
				cursor: 1,
			},
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
	if model.overlay.filterPopup.input != "bc" {
		t.Fatalf("expected edited input bc, got %q", model.overlay.filterPopup.input)
	}
	if model.overlay.filterPopup.cursor != 2 {
		t.Fatalf("expected cursor clamped at input end, got %d", model.overlay.filterPopup.cursor)
	}
}

func TestHandleFilterPopupKey_EscClosesPopup(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active: true,
				step:   filterSelectOperator,
			},
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEsc})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when closing filter popup")
	}
	if model.overlay.filterPopup.active {
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
		listRecords: recordsSpy,
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			tables:   []dto.Table{{Name: "users"}},
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER"},
					{Name: "name", Type: "TEXT"},
				},
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
	if model.overlay.sortPopup.active {
		t.Fatal("expected sort popup to close after apply")
	}
	if model.read.currentSort == nil {
		t.Fatal("expected current sort to be set")
	}
	if model.read.currentSort.Column != "id" {
		t.Fatalf("expected sorted column id, got %q", model.read.currentSort.Column)
	}
	if model.read.currentSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected sort direction DESC, got %s", model.read.currentSort.Direction)
	}
	if recordsSpy.lastSort == nil {
		t.Fatal("expected engine to receive sort")
	}
	if recordsSpy.lastSort.Column != "id" || recordsSpy.lastSort.Direction != dto.SortDirectionDesc {
		t.Fatalf("expected engine sort id DESC, got %+v", recordsSpy.lastSort)
	}
}
