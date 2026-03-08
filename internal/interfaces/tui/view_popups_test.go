package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

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

func TestRenderEditPopup_UsesCombinedSummaryRow(t *testing.T) {
	// Arrange
	model := &Model{
		schema: dto.Schema{
			Columns: []dto.SchemaColumn{
				{
					Name:     "name",
					Type:     "TEXT",
					Nullable: true,
					Input:    dto.ColumnInput{Kind: dto.ColumnInputText},
				},
			},
		},
		editPopup: editPopup{
			active:      true,
			columnIndex: 0,
		},
	}

	// Act
	popup := strings.Join(model.renderEditPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "name (TEXT)"+frameSegmentSeparator+"NULLABLE") {
		t.Fatalf("expected combined summary row for column metadata, got %q", popup)
	}
}

func TestRenderHelpPopup_ShowsOnlyCurrentContextBindings(t *testing.T) {
	// Arrange
	model := &Model{
		height: 40,
		helpPopup: helpPopup{
			active:  true,
			context: helpPopupContextRecords,
		},
	}

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Records: Esc tables") {
		t.Fatalf("expected records context shortcut row in help popup, got %q", popup)
	}
	if !strings.Contains(popup, "w save") {
		t.Fatalf("expected records save shortcut in help popup, got %q", popup)
	}
	if strings.Contains(popup, "Supported Commands") || strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected context-help popup without global help sections, got %q", popup)
	}
}

func TestRenderHelpPopup_UsesConfigPopupHeaderLayout(t *testing.T) {
	// Arrange
	model := &Model{
		height: 40,
		helpPopup: helpPopup{
			active:  true,
			context: helpPopupContextTables,
		},
	}

	// Act
	lines := model.renderHelpPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected help popup to include framed header and content, got %q", strings.Join(lines, "\n"))
	}

	if !strings.HasPrefix(lines[0], frameTopLeft+"Context Help: Tables") {
		t.Fatalf("expected context-specific help title in top border, got %q", lines[0])
	}

	summary := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[1], frameVertical), frameVertical))
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected config-style summary row below title, got %q", summary)
	}

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
}

func TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing(t *testing.T) {
	// Arrange
	model := &Model{
		height:        12,
		helpPopup:     helpPopup{active: true, context: helpPopupContextRecords},
		statusMessage: "",
	}
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	for range 30 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	scrolled := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if strings.Contains(initial, "Shift+S sort") {
		t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
	}
	if !strings.Contains(scrolled, "Shift+S sort") {
		t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
	}
}

func TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow(t *testing.T) {
	// Arrange
	model := &Model{
		height:    12,
		helpPopup: helpPopup{active: true, context: helpPopupContextRecords},
	}
	for range 5 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	before := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyEnter})
	after := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if before != after {
		t.Fatalf("expected non-scroll key to keep help window stable, before=%q after=%q", before, after)
	}
}

func TestView_HelpPopupRendersModalLikeConfigSelector(t *testing.T) {
	// Arrange
	model := &Model{
		width:     80,
		height:    24,
		helpPopup: helpPopup{active: true, context: helpPopupContextTables},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, iconSelection+" Tables") || strings.Contains(view, iconSelection+" Schema") || strings.Contains(view, iconSelection+" Records") {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Context Help: Tables") {
		t.Fatalf("expected help modal frame in view, got %q", view)
	}

	lines := strings.Split(view, "\n")
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Context Help: Tables") {
			helpLine = i
			if strings.Index(line, frameTopLeft+"Context Help: Tables") == 0 {
				t.Fatalf("expected centered modal line with left padding, got %q", line)
			}
			break
		}
	}
	if helpLine <= 0 || helpLine >= len(lines)-1 {
		t.Fatalf("expected help frame to be vertically centered, line=%d total=%d", helpLine, len(lines))
	}
}

func TestView_FilterPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
		filterPopup: filterPopup{
			active: true,
			step:   filterInputValue,
			input:  "abc",
			cursor: 1,
		},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected filter popup modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Filter") {
		t.Fatalf("expected filter popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	filterLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Filter") {
			filterLine = i
			if strings.Index(line, frameTopLeft+"Filter") == 0 {
				t.Fatalf("expected centered filter popup line with left padding, got %q", line)
			}
			break
		}
	}
	if filterLine <= 0 || filterLine >= len(lines)-1 {
		t.Fatalf("expected filter popup to be vertically centered, line=%d total=%d", filterLine, len(lines))
	}
}

func TestView_EditPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
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
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected edit popup modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Edit Cell") {
		t.Fatalf("expected edit popup frame in view, got %q", view)
	}
	lines := strings.Split(view, "\n")
	editLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Edit Cell") {
			editLine = i
			if strings.Index(line, frameTopLeft+"Edit Cell") == 0 {
				t.Fatalf("expected centered edit popup line with left padding, got %q", line)
			}
			break
		}
	}
	if editLine <= 0 || editLine >= len(lines)-1 {
		t.Fatalf("expected edit popup to be vertically centered, line=%d total=%d", editLine, len(lines))
	}
}

func TestView_DirtyConfigPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		viewMode:       ViewRecords,
		width:          80,
		height:         24,
		pendingInserts: []pendingInsertRow{{}},
	}

	// Act
	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "config" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	view := model.View()

	// Assert
	if !strings.Contains(view, frameTopLeft+"Config") {
		t.Fatalf("expected dirty :config modal title, got %q", view)
	}
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected centered modal without background panels, got %q", view)
	}

	lines := strings.Split(view, "\n")
	configLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Config") {
			configLine = i
			if strings.Index(line, frameTopLeft+"Config") == 0 {
				t.Fatalf("expected centered config modal line with left padding, got %q", line)
			}
			break
		}
	}
	if configLine <= 0 || configLine >= len(lines)-1 {
		t.Fatalf("expected config modal to be vertically centered, line=%d total=%d", configLine, len(lines))
	}
}

func TestRenderConfirmPopup_DirtyConfigUsesStandardizedHeaderAndOptionsLayout(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Config",
			message: "Unsaved changes detected. Choose save, discard, or cancel.",
			options: []confirmOption{
				{label: "Save and open config", action: confirmConfigSaveAndOpen},
				{label: "Discard and open config", action: confirmConfigDiscardAndOpen},
				{label: "Cancel", action: confirmConfigCancel},
			},
			selected: 0,
			modal:    true,
		},
	}

	// Act
	lines := model.renderConfirmPopup(60)
	popup := strings.Join(lines, "\n")

	// Assert
	if len(lines) < 6 {
		t.Fatalf("expected modal config popup with framed header and options, got %q", popup)
	}
	if !strings.HasPrefix(lines[0], frameTopLeft+"Config") {
		t.Fatalf("expected config title in top border, got %q", lines[0])
	}
	if !strings.Contains(popup, "Unsaved changes detected.") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
	if !strings.Contains(popup, iconSelection+" Save and open config") {
		t.Fatalf("expected selected option marker in popup, got %q", popup)
	}
}

func TestRenderConfirmPopup_DirtyTableSwitchUsesInformationalMessageAndExplicitActions(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Switch Table",
			message: "Switching tables will cause loss of unsaved data (3 changes). Are you sure you want to discard unsaved data?",
			options: []confirmOption{
				{label: "(y) Yes, discard changes and switch table", action: confirmDiscardTable},
				{label: "(n) No, continue editing", action: confirmCancelTableSwitch},
			},
			selected: 0,
			modal:    true,
		},
	}

	// Act
	lines := model.renderConfirmPopup(120)
	popup := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(popup, frameTopLeft+"Switch Table") {
		t.Fatalf("expected switch-table title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Switching tables will cause loss of unsaved data") {
		t.Fatalf("expected informational switch-table summary in popup, got %q", popup)
	}
	if !strings.Contains(popup, iconSelection+" (y) Yes, discard changes and switch table") {
		t.Fatalf("expected explicit yes action in popup, got %q", popup)
	}
	if !strings.Contains(popup, "(n) No, continue editing") {
		t.Fatalf("expected explicit no action in popup, got %q", popup)
	}
}

func TestView_RegularConfirmPopupRendersAsCenteredModal(t *testing.T) {
	// Arrange
	model := &Model{
		width:  80,
		height: 24,
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Confirm",
			message: "Save staged changes?",
		},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected non-modal confirm as centered popup without background panels, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Confirm") {
		t.Fatalf("expected confirm popup frame, got %q", view)
	}
	lines := strings.Split(view, "\n")
	confirmLine := -1
	for i, line := range lines {
		if strings.Contains(line, frameTopLeft+"Confirm") {
			confirmLine = i
			if strings.Index(line, frameTopLeft+"Confirm") == 0 {
				t.Fatalf("expected centered confirm popup line with left padding, got %q", line)
			}
			break
		}
	}
	if confirmLine <= 0 || confirmLine >= len(lines)-1 {
		t.Fatalf("expected confirm popup to be vertically centered, line=%d total=%d", confirmLine, len(lines))
	}
}

func TestRenderConfirmPopup_InlineUsesStandardizedSeparatorRow(t *testing.T) {
	// Arrange
	model := &Model{
		confirmPopup: confirmPopup{
			active:  true,
			title:   "Confirm",
			message: "Save staged changes?",
		},
	}

	// Act
	lines := model.renderConfirmPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected standardized confirm popup layout, got %q", strings.Join(lines, "\n"))
	}

	separator := strings.TrimSpace(lines[2])
	if !strings.HasPrefix(separator, frameJoinLeft) || !strings.HasSuffix(separator, frameJoinRight) {
		t.Fatalf("expected separator row with border joins, got %q", separator)
	}
	separatorInner := strings.TrimSuffix(strings.TrimPrefix(separator, frameJoinLeft), frameJoinRight)
	if separatorInner == "" || strings.Trim(separatorInner, frameHorizontal) != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
}
