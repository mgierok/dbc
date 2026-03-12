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
	lines := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Name: |") {
		t.Fatalf("expected caret in active name field, got %q", lines)
	}
	if strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Path: |") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Path: |") {
		t.Fatalf("expected caret in active path field after tab, got %q", lines)
	}
	if strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Name: |") {
		t.Fatalf("expected name field to stop being selected after tab, got %q", lines)
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
	lines := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Name: local|") {
		t.Fatalf("expected caret in active edit name field, got %q", lines)
	}
	if strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Path: /tmp/local.sqlite|") {
		t.Fatalf("expected path field to stay inactive before tab, got %q", lines)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyTab})
	lines = stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))

	// Assert
	if !strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Path: /tmp/local.sqlite|") {
		t.Fatalf("expected caret in active edit path field after tab, got %q", lines)
	}
	if strings.Contains(lines, primitives.SelectionSelectedPrefix()+"Name: local|") {
		t.Fatalf("expected name field to stop being selected after tab, got %q", lines)
	}
}

func TestDatabaseSelector_ViewShowsActiveConfigPath(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.json",
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
	if !strings.Contains(view, "/tmp/config.json") {
		t.Fatalf("expected active config path in view, got %q", view)
	}
}

func TestDatabaseSelector_ForcedSetupFormKeepsMainContentFocusedOnFields(t *testing.T) {
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
		t.Fatalf("expected forced setup form main content to stay focused on fields, got %q", view)
	}
	if strings.Contains(view, primitives.SelectorFormSwitchLine()) {
		t.Fatalf("expected forced setup form main content to omit help-only switch guidance, got %q", view)
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

func TestDatabaseSelector_ViewShowsSourceMarkersAndKeepsFooterHint(t *testing.T) {
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
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, primitives.IconConfigSource) {
		t.Fatalf("expected config source marker in selector view, got %q", view)
	}
	if !strings.Contains(view, primitives.IconCLISource) {
		t.Fatalf("expected CLI source marker in selector view, got %q", view)
	}
	if strings.Contains(view, "Legend: "+primitives.IconConfigSource+" config"+primitives.FrameSegmentSeparator+primitives.IconCLISource+" CLI session") {
		t.Fatalf("expected selector main content to stay focused on database rows, got %q", view)
	}
	if !strings.Contains(view, primitives.FrameTopLeft+"Select database") {
		t.Fatalf("expected selector title in top border, got %q", view)
	}
	if !strings.Contains(view, "Context help: ?") {
		t.Fatalf("expected context-help hint in selector, got %q", view)
	}
}

func TestDatabaseSelector_ViewKeepsMainContentFocusedAcrossModes(t *testing.T) {
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
		t.Fatalf("expected browse main content to stay focused on database rows, got %q", browse)
	}

	// Act + Assert: add mode.
	model.openAddForm()
	addView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(addView, primitives.SelectorFormSwitchLine()) || strings.Contains(addView, primitives.SelectorFormSubmitLine("Esc cancel")) {
		t.Fatalf("expected add-form main content to stay focused on editable fields, got %q", addView)
	}

	// Act + Assert: delete-confirm mode.
	model.mode = selectorModeBrowse
	model.openDeleteConfirmation()
	deleteView := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))
	if strings.Contains(deleteView, primitives.SelectorDeleteConfirmationLine()) {
		t.Fatalf("expected delete confirmation main content to stay focused on the selected entry, got %q", deleteView)
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
		t.Fatalf("expected help popup to stay focused on contextual controls, got %q", view)
	}
}
