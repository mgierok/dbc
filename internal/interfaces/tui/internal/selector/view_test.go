package selector

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
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
	if !strings.Contains(lines, primitives.IconSelection+" Name: |") {
		t.Fatalf("expected caret in active name field, got %q", lines)
	}
	if strings.Contains(lines, primitives.IconSelection+" Path: |") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.IconSelection+" Path: |") {
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
	if !strings.Contains(lines, primitives.IconSelection+" Name: local|") {
		t.Fatalf("expected caret in active edit name field, got %q", lines)
	}
	if strings.Contains(lines, primitives.IconSelection+" Path: /tmp/local.sqlite|") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.formLines(), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.IconSelection+" Path: /tmp/local.sqlite|") {
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
	if strings.Contains(view, primitives.SelectorFormSubmitLine("Esc exit app")) {
		t.Fatalf("expected forced setup form to hide shortcut line from main content, got %q", view)
	}
	if strings.Contains(view, primitives.SelectorFormSwitchLine()) {
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
	if !strings.Contains(lines, primitives.IconConfigSource+" local"+primitives.FrameSegmentSeparator+"/tmp/local.sqlite") {
		t.Fatalf("expected config marker in option lines, got %q", lines)
	}
	if !strings.Contains(lines, primitives.IconCLISource+" /tmp/direct.sqlite"+primitives.FrameSegmentSeparator+"/tmp/direct.sqlite") {
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
		if !strings.HasPrefix(trimmed, primitives.FrameTopLeft) &&
			!strings.HasPrefix(trimmed, primitives.FrameVertical) &&
			!strings.HasPrefix(trimmed, primitives.FrameJoinLeft) &&
			!strings.HasPrefix(trimmed, primitives.FrameBottomLeft) {
			continue
		}
		framed := strings.TrimRight(trimmed, " ")
		lineWidth := primitives.TextWidth(framed)
		if !hasFramedLine {
			framedWidth = lineWidth
			hasFramedLine = true
		} else if lineWidth != framedWidth {
			t.Fatalf("expected consistent framed width %d, got %d for line %q", framedWidth, lineWidth, framed)
		}
		if !strings.HasSuffix(framed, primitives.FrameVertical) &&
			!strings.HasSuffix(framed, primitives.FrameTopRight) &&
			!strings.HasSuffix(framed, primitives.FrameJoinRight) &&
			!strings.HasSuffix(framed, primitives.FrameBottomRight) {
			t.Fatalf("expected framed line to end with right border marker, got %q", framed)
		}
	}
	if !hasFramedLine {
		t.Fatalf("expected popup content lines in view, got %q", view)
	}
	if !strings.Contains(view, primitives.IconConfigSource) {
		t.Fatalf("expected config source marker in selector view, got %q", view)
	}
	if !strings.Contains(view, primitives.IconCLISource) {
		t.Fatalf("expected CLI source marker in selector view, got %q", view)
	}
	if strings.Contains(view, "Legend: "+primitives.IconConfigSource+" config"+primitives.FrameSegmentSeparator+primitives.IconCLISource+" CLI session") {
		t.Fatalf("expected legend to be removed from selector main content, got %q", view)
	}
	if !strings.Contains(view, primitives.FrameTopLeft+"Select database") {
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
	if strings.Contains(browse, primitives.SelectorContextLinesBrowseDefault()[0]) || strings.Contains(browse, primitives.SelectorContextLinesBrowseDefault()[1]) {
		t.Fatalf("expected browse shortcuts to be removed from main content, got %q", browse)
	}

	// Act + Assert: add mode.
	model.openAddForm()
	addView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(addView, primitives.SelectorFormSwitchLine()) || strings.Contains(addView, primitives.SelectorFormSubmitLine("Esc cancel")) {
		t.Fatalf("expected add-form shortcuts to be removed from main content, got %q", addView)
	}

	// Act + Assert: delete-confirm mode.
	model.mode = selectorModeBrowse
	model.openDeleteConfirmation()
	deleteView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(deleteView, primitives.SelectorDeleteConfirmationLine()) {
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
	if !strings.Contains(view, primitives.SelectorFormSwitchLine()) || !strings.Contains(view, primitives.SelectorFormSubmitLine("Esc cancel")) {
		t.Fatalf("expected form shortcuts in help popup, got %q", view)
	}
	if strings.Contains(view, primitives.SelectorContextLinesBrowseDefault()[0]) {
		t.Fatalf("expected help popup to exclude browse shortcuts in form context, got %q", view)
	}
	if strings.Contains(view, "Legend: "+primitives.IconConfigSource+" config"+primitives.FrameSegmentSeparator+primitives.IconCLISource+" CLI session") {
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
	model.styles = primitives.NewRenderStyles(true)

	// Act
	lines := model.boxLines(model.listHeight(24), 80)
	view := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(lines[1], "\x1b[1mConfig: /tmp/config.toml\x1b[0m") {
		t.Fatalf("expected selector config path to use popup summary styling, got %q", lines[1])
	}
	selectedLine := lines[3]
	if strings.HasPrefix(selectedLine, "\x1b[7m"+primitives.FrameVertical) {
		t.Fatalf("expected selector left border to remain unstyled, got %q", selectedLine)
	}
	if strings.Contains(selectedLine, primitives.FrameVertical+"\x1b[0m") {
		t.Fatalf("expected selector right border to remain outside selected styling, got %q", selectedLine)
	}
	if !strings.Contains(selectedLine, primitives.FrameVertical+"\x1b[7m "+primitives.SelectionSelectedPrefix()+"⚙ local"+primitives.FrameSegmentSeparator+"/tmp/local.sqlite") {
		t.Fatalf("expected selector selection to style only popup content, got %q", selectedLine)
	}
	if !strings.Contains(stripANSI(lines[len(lines)-2]), "Context help: ?") {
		t.Fatalf("expected selector footer hint in popup footer row, got %q", stripANSI(lines[len(lines)-2]))
	}
	if !strings.Contains(view, "Context help: ?") {
		t.Fatalf("expected selector view to keep context help footer, got %q", view)
	}
}

func TestDatabaseSelector_BoxLines_UsesPopupLikeSelectionStylingForActiveFormField(t *testing.T) {
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
	model.styles = primitives.NewRenderStyles(true)
	model.openAddForm()

	// Act
	lines := model.boxLines(model.listHeight(24), 80)

	// Assert
	var activeLine string
	for _, line := range lines {
		if strings.Contains(line, "Name: |") {
			activeLine = line
			break
		}
	}
	if activeLine == "" {
		t.Fatalf("expected active form field line, got %q", strings.Join(lines, "\n"))
	}
	if strings.HasPrefix(activeLine, "\x1b[7m"+primitives.FrameVertical) {
		t.Fatalf("expected form field left border to remain unstyled, got %q", activeLine)
	}
	if strings.Contains(activeLine, primitives.FrameVertical+"\x1b[0m") {
		t.Fatalf("expected form field right border to remain outside selected styling, got %q", activeLine)
	}
	if !strings.Contains(activeLine, primitives.FrameVertical+"\x1b[7m "+primitives.SelectionSelectedPrefix()+"Name: |") {
		t.Fatalf("expected active form field content to use shared popup selection styling, got %q", activeLine)
	}
}
