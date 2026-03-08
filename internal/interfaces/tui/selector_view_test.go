package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_AddFormShowsCaretInActiveField(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	lines := stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, iconSelection+" Name: |") {
		t.Fatalf("expected caret in active name field, got %q", lines)
	}
	if strings.Contains(lines, iconSelection+" Path: |") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, iconSelection+" Path: |") {
		t.Fatalf("expected caret in active path field after tab, got %q", lines)
	}
}

func TestDatabaseSelector_EditFormShowsCaretInActiveField(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	lines := stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, iconSelection+" Name: local|") {
		t.Fatalf("expected caret in active edit name field, got %q", lines)
	}
	if strings.Contains(lines, iconSelection+" Path: /tmp/local.sqlite|") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, iconSelection+" Path: /tmp/local.sqlite|") {
		t.Fatalf("expected caret in active edit path field after tab, got %q", lines)
	}
}

func TestDatabaseSelector_ViewShowsActiveConfigPath(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.toml",
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.width = 120
	model.height = 24

	// Act
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, "/tmp/config.toml") {
		t.Fatalf("expected active config path in view, got %q", view)
	}
}

func TestDatabaseSelector_ForcedSetupFormHidesShortcutLegendFromMainContent(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	view := strings.Join(model.boxLines(model.listHeight(24), 80), "\n")

	// Assert
	if strings.Contains(view, selectorFormSubmitLine("Esc exit app")) {
		t.Fatalf("expected forced setup form to hide shortcut line from main content, got %q", view)
	}
	if strings.Contains(view, selectorFormSwitchLine()) {
		t.Fatalf("expected forced setup form to hide switch shortcut line from main content, got %q", view)
	}
}

func TestDatabaseSelector_OptionLinesShowSourceMarkers(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	lines := strings.Join(model.optionLines(), "\n")

	// Assert
	if !strings.Contains(lines, iconConfigSource+" local"+frameSegmentSeparator+"/tmp/local.sqlite") {
		t.Fatalf("expected config marker in option lines, got %q", lines)
	}
	if !strings.Contains(lines, iconCLISource+" /tmp/direct.sqlite"+frameSegmentSeparator+"/tmp/direct.sqlite") {
		t.Fatalf("expected CLI marker in option lines, got %q", lines)
	}
}

func TestDatabaseSelector_ViewKeepsRightBorderAlignedWithUnicodeMarkers(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.width = 110
	model.height = 24

	// Act
	view := model.View()
	lines := strings.Split(view, "\n")

	// Assert
	framedWidth := 0
	hasFramedLine := false
	for _, line := range lines {
		trimmed := strings.TrimLeft(stripANSI(line), " ")
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, frameTopLeft) &&
			!strings.HasPrefix(trimmed, frameVertical) &&
			!strings.HasPrefix(trimmed, frameJoinLeft) &&
			!strings.HasPrefix(trimmed, frameBottomLeft) {
			continue
		}
		framed := strings.TrimRight(trimmed, " ")
		lineWidth := textWidth(framed)
		if !hasFramedLine {
			framedWidth = lineWidth
			hasFramedLine = true
		} else if lineWidth != framedWidth {
			t.Fatalf("expected consistent framed width %d, got %d for line %q", framedWidth, lineWidth, framed)
		}
		if !strings.HasSuffix(framed, frameVertical) &&
			!strings.HasSuffix(framed, frameTopRight) &&
			!strings.HasSuffix(framed, frameJoinRight) &&
			!strings.HasSuffix(framed, frameBottomRight) {
			t.Fatalf("expected framed line to end with right border marker, got %q", framed)
		}
	}
	if !hasFramedLine {
		t.Fatalf("expected popup content lines in view, got %q", view)
	}
	if !strings.Contains(view, iconConfigSource) {
		t.Fatalf("expected config source marker in selector view, got %q", view)
	}
	if !strings.Contains(view, iconCLISource) {
		t.Fatalf("expected CLI source marker in selector view, got %q", view)
	}
	if strings.Contains(view, "Legend: "+iconConfigSource+" config"+frameSegmentSeparator+iconCLISource+" CLI session") {
		t.Fatalf("expected legend to be removed from selector main content, got %q", view)
	}
	if !strings.Contains(view, frameTopLeft+"Select database") {
		t.Fatalf("expected selector title in top border, got %q", view)
	}
	if !strings.Contains(view, "Context help: ?") {
		t.Fatalf("expected context-help hint in selector, got %q", view)
	}
}

func TestDatabaseSelector_ViewHidesShortcutLinesAcrossModes(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act + Assert: browse mode.
	browse := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(browse, selectorContextLinesBrowseDefault()[0]) || strings.Contains(browse, selectorContextLinesBrowseDefault()[1]) {
		t.Fatalf("expected browse shortcuts to be removed from main content, got %q", browse)
	}

	// Act + Assert: add mode.
	model.openAddForm()
	addView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(addView, selectorFormSwitchLine()) || strings.Contains(addView, selectorFormSubmitLine("Esc cancel")) {
		t.Fatalf("expected add-form shortcuts to be removed from main content, got %q", addView)
	}

	// Act + Assert: delete-confirm mode.
	model.mode = selectorModeBrowse
	model.openDeleteConfirmation()
	deleteView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(deleteView, selectorDeleteConfirmationLine()) {
		t.Fatalf("expected delete-confirm shortcuts to be removed from main content, got %q", deleteView)
	}
}

func TestDatabaseSelector_BoxLinesEnforceMinimumHeight40Percent(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.height = 50

	// Act
	lines := model.boxLines(model.listHeight(model.height), 100)

	// Assert
	minExpectedHeight := (model.height*40 + 99) / 100
	if len(lines) < minExpectedHeight {
		t.Fatalf("expected selector min height %d, got %d", minExpectedHeight, len(lines))
	}
}

func TestDatabaseSelector_ContextHelpPopupRendersCurrentModeShortcutsOnly(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.openAddForm()
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Act
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, "Context Help: Config") {
		t.Fatalf("expected selector help title, got %q", view)
	}
	if !strings.Contains(view, selectorFormSwitchLine()) || !strings.Contains(view, selectorFormSubmitLine("Esc cancel")) {
		t.Fatalf("expected form shortcuts in help popup, got %q", view)
	}
	if strings.Contains(view, selectorContextLinesBrowseDefault()[0]) {
		t.Fatalf("expected help popup to exclude browse shortcuts in form context, got %q", view)
	}
	if strings.Contains(view, "Legend: "+iconConfigSource+" config"+frameSegmentSeparator+iconCLISource+" CLI session") {
		t.Fatalf("expected help popup to exclude legend, got %q", view)
	}
}

func TestDatabaseSelector_BoxLines_UsesPopupLikeSummaryAndSelectionStyling(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.toml",
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.styles = renderStyles{enabled: true}

	// Act
	lines := model.boxLines(model.listHeight(24), 80)
	view := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(lines[1], "\x1b[1mConfig: /tmp/config.toml\x1b[0m") {
		t.Fatalf("expected selector config path to use popup summary styling, got %q", lines[1])
	}
	if !strings.Contains(view, "\x1b[7m"+frameVertical+" "+selectionSelectedPrefix()+"⚙ local"+frameSegmentSeparator+"/tmp/local.sqlite") {
		t.Fatalf("expected selector selection to use popup-like full-line styling, got %q", view)
	}
}
