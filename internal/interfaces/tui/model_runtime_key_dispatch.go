package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ui.saveInFlight || m.ui.runtimeSwitchInFlight {
		return m, nil
	}

	if m.overlay.databaseSelector.active {
		return m.handleRuntimeDatabaseSelectorKey(msg)
	}

	key := msg.String()

	if primitives.KeyMatches(primitives.KeyRuntimeOpenContextHelp, key) {
		if m.overlay.helpPopup.active {
			return m, nil
		}
		m.openHelpPopup(m.currentHelpPopupContext())
		return m, nil
	}
	if primitives.KeyMatches(primitives.KeyRuntimeOpenCommandInput, key) &&
		!m.overlay.commandInput.active &&
		m.commandInputSupportedInCurrentContext() {
		return m.startCommandInput()
	}

	if m.overlay.helpPopup.active {
		return m.handleHelpPopupKey(msg)
	}
	if m.overlay.editPopup.active {
		return m.handleEditPopupKey(msg)
	}
	if m.overlay.confirmPopup.active {
		return m.handleConfirmPopupKey(msg)
	}
	if m.overlay.filterPopup.active {
		return m.handleFilterPopupKey(msg)
	}
	if m.overlay.sortPopup.active {
		return m.handleSortPopupKey(msg)
	}
	if m.overlay.commandInput.active {
		return m.handleCommandInputKey(msg)
	}
	if m.overlay.recordDetail.active {
		return m.handleRecordDetailKey(msg)
	}

	if m.overlay.pendingG {
		if primitives.KeyMatches(primitives.KeyRuntimeJumpTopPending, key) {
			m.overlay.pendingG = false
			return m.jumpTop()
		}
		m.overlay.pendingG = false
	}

	switch {
	case primitives.KeyMatches(primitives.KeyRuntimeJumpTopPending, key):
		m.overlay.pendingG = true
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeJumpBottom, key):
		return m.jumpBottom()
	case primitives.KeyMatches(primitives.KeyRuntimeEnter, key):
		if m.read.viewMode == ViewRecords && m.read.focus == FocusContent {
			return m.openRecordDetail()
		}
		return m.switchToRecords()
	case primitives.KeyMatches(primitives.KeyRuntimeEdit, key):
		if m.read.viewMode == ViewRecords && m.read.focus == FocusContent {
			if !m.read.recordFieldFocus {
				return m.enableRecordFieldFocus()
			}
			return m.openEditPopup()
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeEsc, key):
		if m.read.viewMode == ViewRecords && m.read.recordFieldFocus {
			m.read.recordFieldFocus = false
			return m, nil
		}
		if m.read.focus == FocusContent {
			m.read.focus = FocusTables
			m.read.viewMode = ViewSchema
			return m, nil
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeyRuntimeFilter, key):
		return m.startFilterPopup()
	case primitives.KeyMatches(primitives.KeyRuntimeSort, key):
		return m.startSortPopup()
	case primitives.KeyMatches(primitives.KeyRuntimeRecordDetail, key):
		return m.openRecordDetail()
	case primitives.KeyMatches(primitives.KeyRuntimeInsert, key):
		return m.addPendingInsert()
	case primitives.KeyMatches(primitives.KeyRuntimeDelete, key):
		return m.toggleDeleteSelection()
	case primitives.KeyMatches(primitives.KeyRuntimeUndo, key):
		return m.undoStagedAction()
	case primitives.KeyMatches(primitives.KeyRuntimeRedo, key):
		return m.redoStagedAction()
	case primitives.KeyMatches(primitives.KeyRuntimeToggleAutoFields, key):
		return m.toggleInsertAutoFields()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveDown, key):
		return m.moveDown()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveUp, key):
		return m.moveUp()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveLeft, key):
		return m.moveLeft()
	case primitives.KeyMatches(primitives.KeyRuntimeMoveRight, key):
		return m.moveRight()
	case primitives.KeyMatches(primitives.KeyRuntimePageDown, key):
		return m.pageDown()
	case primitives.KeyMatches(primitives.KeyRuntimePageUp, key):
		return m.pageUp()
	default:
		return m, nil
	}
}
