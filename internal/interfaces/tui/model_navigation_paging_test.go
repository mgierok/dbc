package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestHandleKey_FieldFocusNavigationAdjustsColumnForPendingInsertRows(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordFieldFocus: true,
		recordSelection:  1,
		recordColumn:     0,
		staging: stagingState{
			pendingInserts: []pendingInsertRow{
				{
					values: map[int]stagedEdit{
						0: {Value: dto.StagedValue{Text: "", Raw: ""}},
						1: {Value: dto.StagedValue{Text: "new", Raw: "new"}},
					},
					explicitAuto: map[int]bool{},
				},
			},
		},
		records: []dto.RecordRow{
			{Values: []string{"1", "alice"}},
		},
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "TEXT", Nullable: false},
			},
		},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

	// Assert
	if model.recordSelection != 0 {
		t.Fatalf("expected selection to move to pending insert row, got %d", model.recordSelection)
	}
	if model.recordColumn != 1 {
		t.Fatalf("expected focused column to move off hidden auto-increment field, got %d", model.recordColumn)
	}
}

func TestHandleKey_CtrlFInRecordsLoadsNextPage(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"21", "alice"}},
			},
			TotalCount: 45,
		},
	}
	model := &Model{
		ctx:              context.Background(),
		viewMode:         ViewRecords,
		focus:            FocusContent,
		listRecords:      recordsSpy,
		tables:           []dto.Table{{Name: "users"}},
		recordPageIndex:  0,
		recordTotalPages: 3,
		recordTotalCount: 45,
	}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	if cmd == nil {
		t.Fatal("expected command to load next page")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.recordPageIndex != 1 {
		t.Fatalf("expected current page index 1, got %d", model.recordPageIndex)
	}
	if recordsSpy.lastRecordsOffset != 20 {
		t.Fatalf("expected offset 20 for second page, got %d", recordsSpy.lastRecordsOffset)
	}
	if recordsSpy.lastRecordsLimit != 20 {
		t.Fatalf("expected limit 20, got %d", recordsSpy.lastRecordsLimit)
	}
	if model.recordTotalPages != 3 {
		t.Fatalf("expected 3 pages, got %d", model.recordTotalPages)
	}
}

func TestHandleKey_CtrlBInRecordsLoadsPreviousPage(t *testing.T) {
	// Arrange
	recordsSpy := &spyListRecordsUseCase{
		page: dto.RecordPage{
			Rows: []dto.RecordRow{
				{Values: []string{"1", "alice"}},
			},
			TotalCount: 45,
		},
	}
	model := &Model{
		ctx:              context.Background(),
		viewMode:         ViewRecords,
		focus:            FocusContent,
		listRecords:      recordsSpy,
		tables:           []dto.Table{{Name: "users"}},
		recordPageIndex:  1,
		recordTotalPages: 3,
		recordTotalCount: 45,
	}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlB})
	if cmd == nil {
		t.Fatal("expected command to load previous page")
	}
	msg := cmd()
	model.Update(msg)

	// Assert
	if model.recordPageIndex != 0 {
		t.Fatalf("expected current page index 0, got %d", model.recordPageIndex)
	}
	if recordsSpy.lastRecordsOffset != 0 {
		t.Fatalf("expected offset 0 for first page, got %d", recordsSpy.lastRecordsOffset)
	}
	if recordsSpy.lastRecordsLimit != 20 {
		t.Fatalf("expected limit 20, got %d", recordsSpy.lastRecordsLimit)
	}
}

func TestHandleKey_CtrlBDoesNotGoBeforeFirstPage(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordPageIndex:  0,
		recordTotalPages: 3,
	}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlB})

	// Assert
	if cmd != nil {
		t.Fatal("expected no load command on first page")
	}
	if model.recordPageIndex != 0 {
		t.Fatalf("expected to stay on first page, got %d", model.recordPageIndex)
	}
}

func TestHandleKey_CtrlFDoesNotGoBeyondLastPage(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:         ViewRecords,
		focus:            FocusContent,
		recordPageIndex:  2,
		recordTotalPages: 3,
	}

	// Act
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlF})

	// Assert
	if cmd != nil {
		t.Fatal("expected no load command on last page")
	}
	if model.recordPageIndex != 2 {
		t.Fatalf("expected to stay on last page, got %d", model.recordPageIndex)
	}
}
