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

func TestDatabaseSelector_EditFormSanitizesPreloadedValuesAndPreservesCaret(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "lo\x1b[31m\r\ncal", Path: "/tmp/db\t.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	// Act
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	view := stripANSI(strings.Join(model.boxLines(model.listHeight(24), 80), "\n"))

	// Assert
	if strings.Contains(view, "\x1b") || strings.Contains(view, "\r") || strings.Contains(view, "\n\t") || strings.Contains(view, "\t") {
		t.Fatalf("expected sanitized selector edit form output, got %q", view)
	}
	if !strings.Contains(view, primitives.SelectionSelectedPrefix()+"Name: lo cal|") {
		t.Fatalf("expected sanitized name with preserved caret, got %q", view)
	}
	if !strings.Contains(view, "Path: /tmp/db .sqlite") {
		t.Fatalf("expected sanitized path value, got %q", view)
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
	if strings.Contains(browse, primitives.SelectorContextLinesBrowseDefault("quit")[0]) || strings.Contains(browse, primitives.SelectorContextLinesBrowseDefault("quit")[1]) {
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
	if strings.Contains(view, primitives.SelectorContextLinesBrowseDefault("quit")[0]) {
		t.Fatalf("expected help popup to exclude browse shortcuts in form context, got %q", view)
	}
	if strings.Contains(view, "Legend: "+primitives.IconConfigSource+" config"+primitives.FrameSegmentSeparator+primitives.IconCLISource+" CLI session") {
		t.Fatalf("expected help popup to stay focused on contextual controls, got %q", view)
	}
}

func TestDatabaseSelector_ContextHelpPopupUsesRuntimeResumeBrowseEscLabel(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		BrowseEscBehavior: SelectorBrowseEscBehaviorRuntimeResume,
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model = sendKey(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	// Act
	view := stripANSI(model.View())

	// Assert
	if !strings.Contains(view, "Esc close") {
		t.Fatalf("expected runtime-resume help label, got %q", view)
	}
	if strings.Contains(view, "Esc quit") {
		t.Fatalf("expected runtime-resume help to omit startup quit label, got %q", view)
	}
}

func TestDatabaseSelector_ViewUsesSemanticRolesForFormLabelsAndErrors(t *testing.T) {
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
	model.form.errorMessage = "invalid path"

	// Act
	view := strings.Join(model.boxLines(model.listHeight(24), 80), "\n")

	// Assert
	if !strings.Contains(view, "\x1b[1mPath:\x1b[0m") {
		t.Fatalf("expected semantic label styling for form field, got %q", view)
	}
	if !strings.Contains(view, "\x1b[1;4mError: invalid path\x1b[0m") {
		t.Fatalf("expected semantic error styling for form error, got %q", view)
	}
}

func TestDatabaseSelector_ViewSanitizesBrowseStatusAndDeleteConfirmationContent(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.json\x1b[31m",
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "ana\x1b[31mlytics",
				ConnString: "/tmp/db\r\n.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	model.width = 100
	model.height = 24
	model.browse.selected = 1
	model.browse.statusMessage = "Delete failed: boom\x1b]2;ignored\awhy"

	// Act
	browseView := stripANSI(model.View())

	// Assert
	if strings.Contains(browseView, "\x1b") || strings.Contains(browseView, "\r") {
		t.Fatalf("expected sanitized selector browse output, got %q", browseView)
	}
	if !strings.Contains(browseView, "/tmp/config.json") {
		t.Fatalf("expected sanitized config path, got %q", browseView)
	}
	if !strings.Contains(browseView, "analytics") {
		t.Fatalf("expected sanitized option name, got %q", browseView)
	}
	if !strings.Contains(browseView, "/tmp/db .sqlite") {
		t.Fatalf("expected sanitized option path, got %q", browseView)
	}
	if !strings.Contains(browseView, "Delete failed: boomwhy") {
		t.Fatalf("expected sanitized browse status, got %q", browseView)
	}

	// Act
	model.openDeleteConfirmation()
	deleteView := stripANSI(model.View())

	// Assert
	if !strings.Contains(deleteView, "analytics"+primitives.FrameSegmentSeparator+"/tmp/db .sqlite") {
		t.Fatalf("expected sanitized delete confirmation content, got %q", deleteView)
	}
}
