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

func TestHandleFilterPopupKey_EnterLoadsOperatorsForSelectedColumn(t *testing.T) {
	// Arrange
	operatorsSpy := &spyListOperatorsUseCase{
		operators: []dto.Operator{
			{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		},
	}
	model := &Model{
		ctx:           context.Background(),
		listOperators: operatorsSpy,
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
	if len(model.overlay.filterPopup.operators) != 1 {
		t.Fatalf("expected one loaded operator, got %d", len(model.overlay.filterPopup.operators))
	}
}

func TestHandleFilterPopupKey_EnterMovesToValueInputWhenOperatorRequiresValue(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active:        true,
				step:          filterSelectOperator,
				operatorIndex: 0,
				operators: []dto.Operator{
					{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				},
			},
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving to value-input step")
	}
	if model.overlay.filterPopup.step != filterInputValue {
		t.Fatalf("expected value-input step, got %v", model.overlay.filterPopup.step)
	}
	if model.overlay.filterPopup.input != "" {
		t.Fatalf("expected empty input when opening value step, got %q", model.overlay.filterPopup.input)
	}
	if model.overlay.filterPopup.cursor != 0 {
		t.Fatalf("expected cursor reset to 0, got %d", model.overlay.filterPopup.cursor)
	}
}

func TestHandleFilterPopupKey_EnterAppliesFilterWithoutValueWhenOperatorDoesNotRequireIt(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{}
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
			recordPageIndex: 2,
		},
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active:        true,
				step:          filterSelectOperator,
				columnIndex:   1,
				operatorIndex: 0,
				operators: []dto.Operator{
					{Name: "Is Null", Kind: dto.OperatorKindIsNull, RequiresValue: false},
				},
			},
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after filter apply")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to close after apply")
	}
	assertFilterEqual(t, model.read.currentFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Is Null", Kind: dto.OperatorKindIsNull, RequiresValue: false},
		Value:    "",
	})
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after filter apply, got %d", model.read.recordPageIndex)
	}
	assertFilterEqual(t, recordsSpy.lastFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Is Null", Kind: dto.OperatorKindIsNull, RequiresValue: false},
		Value:    "",
	})
}

func TestHandleFilterPopupKey_EnterAppliesFilterWithCurrentInputValue(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{}
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
			recordPageIndex: 3,
		},
		overlay: runtimeOverlayState{
			filterPopup: filterPopup{
				active:        true,
				step:          filterInputValue,
				columnIndex:   1,
				operatorIndex: 0,
				input:         "alice",
				operators: []dto.Operator{
					{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
				},
			},
		},
	}

	// Act
	_, cmd := model.handleFilterPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected records reload command after filter apply")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.overlay.filterPopup.active {
		t.Fatal("expected filter popup to close after apply")
	}
	assertFilterEqual(t, model.read.currentFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		Value:    "alice",
	})
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after filter apply, got %d", model.read.recordPageIndex)
	}
	assertFilterEqual(t, recordsSpy.lastFilter, &dto.Filter{
		Column:   "name",
		Operator: dto.Operator{Name: "Equals", Kind: dto.OperatorKindEq, RequiresValue: true},
		Value:    "alice",
	})
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

func TestHandleSortPopupKey_EnterMovesToDirectionStep(t *testing.T) {
	// Arrange
	model := &Model{
		overlay: runtimeOverlayState{
			sortPopup: sortPopup{
				active: true,
				step:   sortSelectColumn,
			},
		},
	}

	// Act
	_, cmd := model.handleSortPopupKey(tea.KeyMsg{Type: tea.KeyEnter})

	// Assert
	if cmd != nil {
		t.Fatal("expected no command when moving to direction step")
	}
	if model.overlay.sortPopup.step != sortSelectDirection {
		t.Fatalf("expected direction step, got %v", model.overlay.sortPopup.step)
	}
}

func TestHandleSortPopupKey_EnterAppliesSelectedSort(t *testing.T) {
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
			recordPageIndex: 4,
		},
		overlay: runtimeOverlayState{
			sortPopup: sortPopup{
				active:         true,
				step:           sortSelectDirection,
				columnIndex:    0,
				directionIndex: 1,
			},
		},
	}

	// Act
	_, cmd := model.handleSortPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
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
	assertSortEqual(t, model.read.currentSort, &dto.Sort{
		Column:    "id",
		Direction: dto.SortDirectionDesc,
	})
	if model.read.recordPageIndex != 0 {
		t.Fatalf("expected page index reset to 0 after sort apply, got %d", model.read.recordPageIndex)
	}
	assertSortEqual(t, recordsSpy.lastSort, &dto.Sort{
		Column:    "id",
		Direction: dto.SortDirectionDesc,
	})
}
