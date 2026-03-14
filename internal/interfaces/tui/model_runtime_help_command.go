package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) startCommandInput() (tea.Model, tea.Cmd) {
	m.clearPendingRuntimeKeyState()
	m.overlay.commandInput = commandInput{
		active: true,
		value:  "",
		cursor: 0,
	}
	m.ui.statusMessage = ""
	return m, nil
}

func (m *Model) clearPendingRuntimeKeyState() {
	m.overlay.pendingG = false
}

func (m *Model) submitCommandInput() (tea.Model, tea.Cmd) {
	command := ":" + strings.TrimSpace(m.overlay.commandInput.value)
	m.overlay.commandInput = commandInput{}

	commandSpec, err := primitives.ParseRuntimeCommand(command)
	if err != nil {
		if primitives.IsUnknownRuntimeCommand(err) {
			m.ui.statusMessage = fmt.Sprintf("Unknown command: %s", command)
			return m, nil
		}
		m.ui.statusMessage = "Error: " + err.Error()
		return m, nil
	}

	switch commandSpec.Action {
	case primitives.RuntimeCommandActionSetRecordLimit:
		return m.applyRecordLimit(commandSpec.RecordLimit)
	case primitives.RuntimeCommandActionOpenHelp:
		m.openHelpPopup(m.currentHelpPopupContext())
		return m, nil
	case primitives.RuntimeCommandActionQuit:
		if m.hasDirtyEdits() {
			prompt := m.dirtyNavigationPolicyUseCase().BuildQuitPrompt(m.dirtyEditCount())
			m.openModalConfirmPopupWithOptions(
				prompt.Title,
				prompt.Message,
				m.confirmOptionsFromDirtyPrompt(prompt, dirtyConfirmFlowQuit),
				0,
			)
			return m, nil
		}
		return m, tea.Quit
	case primitives.RuntimeCommandActionOpenConfig:
		if m.hasDirtyEdits() {
			prompt := m.dirtyNavigationPolicyUseCase().BuildConfigPrompt()
			m.openModalConfirmPopupWithOptions(
				prompt.Title,
				prompt.Message,
				m.confirmOptionsFromDirtyPrompt(prompt, dirtyConfirmFlowConfig),
				0,
			)
			return m, nil
		}
		m.ui.openConfigSelector = true
		m.ui.statusMessage = "Opening database selector"
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) openHelpPopup(context helpPopupContext) {
	m.overlay.commandInput = commandInput{}
	m.overlay.helpPopup = helpPopup{
		active:       true,
		scrollOffset: 0,
		context:      context,
	}
}

func (m *Model) closeHelpPopup() {
	m.overlay.helpPopup = helpPopup{}
}

func (m *Model) moveHelpPopupScroll(delta int) {
	maxOffset := m.helpPopupMaxOffset()
	m.overlay.helpPopup.scrollOffset = clamp(m.overlay.helpPopup.scrollOffset+delta, 0, maxOffset)
}

func (m *Model) helpPopupVisibleLines() int {
	const minVisibleLines = 6
	const maxVisibleLines = 12

	visible := m.contentHeight() - 10
	if visible < minVisibleLines {
		return minVisibleLines
	}
	if visible > maxVisibleLines {
		return maxVisibleLines
	}
	return visible
}

func (m *Model) helpPopupMaxOffset() int {
	maxOffset := len(m.helpPopupContentLines()) - m.helpPopupVisibleLines()
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (m *Model) currentHelpPopupContext() helpPopupContext {
	switch {
	case m.overlay.editPopup.active:
		return helpPopupContextEditPopup
	case m.overlay.confirmPopup.active:
		return helpPopupContextConfirmPopup
	case m.overlay.filterPopup.active:
		return helpPopupContextFilterPopup
	case m.overlay.sortPopup.active:
		return helpPopupContextSortPopup
	case m.overlay.helpPopup.active:
		return helpPopupContextHelpPopup
	case m.overlay.commandInput.active:
		return helpPopupContextCommandInput
	case m.overlay.recordDetail.active:
		return helpPopupContextRecordDetail
	case m.read.focus == FocusTables:
		return helpPopupContextTables
	case m.read.focus == FocusContent && m.read.viewMode == ViewSchema:
		return helpPopupContextSchema
	case m.read.focus == FocusContent && m.read.viewMode == ViewRecords:
		return helpPopupContextRecords
	default:
		return helpPopupContextUnknown
	}
}

func (m *Model) helpPopupContextTitle() string {
	switch m.overlay.helpPopup.context {
	case helpPopupContextTables:
		return "Context Help: Tables"
	case helpPopupContextSchema:
		return "Context Help: Schema"
	case helpPopupContextRecords:
		return "Context Help: Records"
	case helpPopupContextRecordDetail:
		return "Context Help: Record Detail"
	case helpPopupContextFilterPopup:
		return "Context Help: Filter Popup"
	case helpPopupContextSortPopup:
		return "Context Help: Sort Popup"
	case helpPopupContextEditPopup:
		return "Context Help: Edit Popup"
	case helpPopupContextConfirmPopup:
		return "Context Help: Confirm Popup"
	case helpPopupContextCommandInput:
		return "Context Help: Command Input"
	case helpPopupContextHelpPopup:
		return "Context Help: Help Popup"
	default:
		return "Context Help"
	}
}

func (m *Model) helpPopupContentLines() []string {
	shortcuts := m.helpPopupContextShortcuts()
	if strings.TrimSpace(shortcuts) == "" {
		return []string{"No keybindings available in this context."}
	}
	parts := strings.Split(shortcuts, primitives.FrameSegmentSeparator)
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return []string{"No keybindings available in this context."}
	}
	return lines
}

func (m *Model) helpPopupContextShortcuts() string {
	switch m.overlay.helpPopup.context {
	case helpPopupContextEditPopup:
		return primitives.RuntimeStatusEditShortcuts()
	case helpPopupContextConfirmPopup:
		return primitives.RuntimeStatusConfirmShortcuts(len(m.overlay.confirmPopup.options) > 0)
	case helpPopupContextFilterPopup:
		return primitives.RuntimeStatusFilterPopupShortcuts()
	case helpPopupContextSortPopup:
		return primitives.RuntimeStatusSortPopupShortcuts()
	case helpPopupContextHelpPopup:
		return primitives.RuntimeStatusHelpPopupShortcuts()
	case helpPopupContextCommandInput:
		return primitives.RuntimeStatusCommandInputShortcuts()
	case helpPopupContextRecordDetail:
		return primitives.RuntimeStatusRecordDetailShortcuts()
	case helpPopupContextTables:
		return primitives.RuntimeStatusTablesShortcuts()
	case helpPopupContextSchema:
		return primitives.RuntimeStatusSchemaShortcuts()
	case helpPopupContextRecords:
		return primitives.RuntimeStatusRecordsShortcuts()
	default:
		return ""
	}
}

func (m *Model) handleHelpPopupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.closeHelpPopup()
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveDown, key):
		m.moveHelpPopupScroll(1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupMoveUp, key):
		m.moveHelpPopupScroll(-1)
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		m.moveHelpPopupScroll(m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		m.moveHelpPopupScroll(-m.helpPopupVisibleLines())
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpTop, key):
		m.overlay.helpPopup.scrollOffset = 0
		return m, nil
	case primitives.KeyMatches(primitives.KeyPopupJumpBottom, key):
		m.overlay.helpPopup.scrollOffset = m.helpPopupMaxOffset()
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) handleCommandInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		m.overlay.commandInput = commandInput{}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		return m.submitCommandInput()
	case primitives.KeyMatches(primitives.KeyInputMoveLeft, key):
		m.overlay.commandInput.cursor = clamp(m.overlay.commandInput.cursor-1, 0, len(m.overlay.commandInput.value))
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputMoveRight, key):
		m.overlay.commandInput.cursor = clamp(m.overlay.commandInput.cursor+1, 0, len(m.overlay.commandInput.value))
		return m, nil
	case primitives.KeyMatches(primitives.KeyInputBackspace, key):
		if m.overlay.commandInput.value != "" {
			m.overlay.commandInput.value, m.overlay.commandInput.cursor = deleteAtCursor(m.overlay.commandInput.value, m.overlay.commandInput.cursor)
		}
		return m, nil
	}

	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		insert := string(msg.Runes)
		if msg.Type == tea.KeySpace {
			insert = " "
		}
		m.overlay.commandInput.value, m.overlay.commandInput.cursor = insertAtCursor(m.overlay.commandInput.value, insert, m.overlay.commandInput.cursor)
	}
	return m, nil
}
