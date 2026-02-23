package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	domainmodel "github.com/mgierok/dbc/internal/domain/model"
)

func TestStatusShortcuts_TablesPanel(t *testing.T) {
	// Arrange
	model := &Model{focus: FocusTables}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Tables: Enter records | F filter" {
		t.Fatalf("expected table shortcuts, got %q", shortcuts)
	}
}

func TestStatusShortcuts_SchemaPanel(t *testing.T) {
	// Arrange
	model := &Model{
		focus:    FocusContent,
		viewMode: ViewSchema,
	}

	// Act
	shortcuts := model.statusShortcuts()

	// Assert
	if shortcuts != "Schema: Esc tables | F filter" {
		t.Fatalf("expected schema shortcuts, got %q", shortcuts)
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
	if shortcuts != "Records: Esc tables | Enter edit | i insert | d delete | u undo | Ctrl+r redo | w save | F filter" {
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

func TestRenderHelpPopup_IncludesRequiredSectionsAndOneLineDescriptions(t *testing.T) {
	// Arrange
	model := &Model{
		height:    40,
		helpPopup: helpPopup{active: true},
	}

	// Act
	popup := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if !strings.Contains(popup, "Supported Commands") {
		t.Fatalf("expected help popup to include Supported Commands section, got %q", popup)
	}
	if !strings.Contains(popup, "Supported Keywords") {
		t.Fatalf("expected help popup to include Supported Keywords section, got %q", popup)
	}
	if !strings.Contains(popup, ":config / :c - Open database selector and config manager.") {
		t.Fatalf("expected help popup to include :config alias one-line description, got %q", popup)
	}
	if !strings.Contains(popup, ":help / :h - Open runtime help popup reference.") {
		t.Fatalf("expected help popup to include :help alias one-line description, got %q", popup)
	}
	if !strings.Contains(popup, ":q / :quit - Quit the application.") {
		t.Fatalf("expected help popup to include :q alias one-line description, got %q", popup)
	}
	if strings.Contains(popup, "q / Ctrl+c - Quit the application.") {
		t.Fatalf("expected help popup to avoid runtime q/Ctrl+c quit shortcut, got %q", popup)
	}
	if !strings.Contains(popup, "Esc - Close active popup/context.") {
		t.Fatalf("expected help popup to include Esc one-line description, got %q", popup)
	}
}

func TestRenderHelpPopup_UsesConfigPopupHeaderLayout(t *testing.T) {
	// Arrange
	model := &Model{
		height:    40,
		helpPopup: helpPopup{active: true},
	}

	// Act
	lines := model.renderHelpPopup(60)

	// Assert
	if len(lines) < 5 {
		t.Fatalf("expected help popup to include framed header and content, got %q", strings.Join(lines, "\n"))
	}

	summary := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[2], "|"), "|"))
	if summary != "Use j/k, Ctrl+f/Ctrl+b to scroll. Esc closes." {
		t.Fatalf("expected config-style summary row below title, got %q", summary)
	}

	separator := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[3], "|"), "|"))
	if separator == "" || strings.Trim(separator, "-") != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
}

func TestRenderHelpPopup_ScrollCanReachFinalItemWhenOverflowing(t *testing.T) {
	// Arrange
	model := &Model{
		height:        12,
		helpPopup:     helpPopup{active: true},
		statusMessage: "",
	}
	initial := strings.Join(model.renderHelpPopup(60), "\n")

	// Act
	for range 30 {
		model.handleHelpPopupKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	scrolled := strings.Join(model.renderHelpPopup(60), "\n")

	// Assert
	if strings.Contains(initial, "Ctrl+a - Toggle auto field visibility for inserts.") {
		t.Fatalf("expected final help item to be hidden before scrolling, got %q", initial)
	}
	if !strings.Contains(scrolled, "Ctrl+a - Toggle auto field visibility for inserts.") {
		t.Fatalf("expected final help item to be reachable after scrolling, got %q", scrolled)
	}
}

func TestHandleHelpPopupKey_NonScrollKeyDoesNotChangeRenderedWindow(t *testing.T) {
	// Arrange
	model := &Model{
		height:    12,
		helpPopup: helpPopup{active: true},
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
		helpPopup: helpPopup{active: true},
	}

	// Act
	view := model.View()

	// Assert
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected help modal view without background panels, got %q", view)
	}
	if !strings.Contains(view, "|Help") {
		t.Fatalf("expected help modal frame in view, got %q", view)
	}

	lines := strings.Split(view, "\n")
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Help") {
			helpLine = i
			if strings.Index(line, "|Help") == 0 {
				t.Fatalf("expected centered modal line with left padding, got %q", line)
			}
			break
		}
	}
	if helpLine <= 0 || helpLine >= len(lines)-1 {
		t.Fatalf("expected help frame to be vertically centered, line=%d total=%d", helpLine, len(lines))
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
	if !strings.Contains(view, "|Config") {
		t.Fatalf("expected dirty :config modal title, got %q", view)
	}
	if strings.Contains(view, "Tables") || strings.Contains(view, "Schema") || strings.Contains(view, "Records") {
		t.Fatalf("expected centered modal without background panels, got %q", view)
	}

	lines := strings.Split(view, "\n")
	configLine := -1
	for i, line := range lines {
		if strings.Contains(line, "|Config") {
			configLine = i
			if strings.Index(line, "|Config") == 0 {
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
	if !strings.Contains(popup, "|Config") {
		t.Fatalf("expected config title in popup, got %q", popup)
	}
	if !strings.Contains(popup, "Unsaved changes detected. Choose save, discard, or cancel.") {
		t.Fatalf("expected decision summary in popup, got %q", popup)
	}

	separator := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lines[3], "|"), "|"))
	if separator == "" || strings.Trim(separator, "-") != "" {
		t.Fatalf("expected separator row after summary, got %q", separator)
	}
	if !strings.Contains(popup, "> Save and open config") {
		t.Fatalf("expected selected option marker in popup, got %q", popup)
	}
}

func TestView_RegularConfirmPopupStaysInlineWithPanels(t *testing.T) {
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
	if !strings.Contains(view, "Tables") {
		t.Fatalf("expected background panels for non-modal confirm, got %q", view)
	}
	if !strings.Contains(view, "|Confirm") {
		t.Fatalf("expected inline confirm popup frame, got %q", view)
	}
}
