package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if keyMatches(keyRuntimeOpenContextHelp, key) {
		if m.helpPopup.active {
			return m, nil
		}
		m.openHelpPopup(m.currentHelpPopupContext())
		return m, nil
	}

	if m.helpPopup.active {
		return m.handleHelpPopupKey(msg)
	}
	if m.editPopup.active {
		return m.handleEditPopupKey(msg)
	}
	if m.confirmPopup.active {
		return m.handleConfirmPopupKey(msg)
	}
	if m.filterPopup.active {
		return m.handleFilterPopupKey(msg)
	}
	if m.sortPopup.active {
		return m.handleSortPopupKey(msg)
	}
	if m.commandInput.active {
		return m.handleCommandInputKey(msg)
	}
	if m.recordDetail.active {
		return m.handleRecordDetailKey(msg)
	}

	if m.pendingG {
		if keyMatches(keyRuntimeJumpTopPending, key) {
			m.pendingG = false
			return m.jumpTop()
		}
		m.pendingG = false
	}

	switch {
	case keyMatches(keyRuntimeOpenCommandInput, key):
		return m.startCommandInput()
	case keyMatches(keyRuntimeJumpTopPending, key):
		m.pendingG = true
		return m, nil
	case keyMatches(keyRuntimeJumpBottom, key):
		return m.jumpBottom()
	case keyMatches(keyRuntimeEnter, key):
		if m.viewMode == ViewRecords && m.focus == FocusContent {
			return m.openRecordDetail()
		}
		return m.switchToRecords()
	case keyMatches(keyRuntimeEdit, key):
		if m.viewMode == ViewRecords && m.focus == FocusContent {
			if !m.recordFieldFocus {
				return m.enableRecordFieldFocus()
			}
			return m.openEditPopup()
		}
		return m, nil
	case keyMatches(keyRuntimeEsc, key):
		if m.viewMode == ViewRecords && m.recordFieldFocus {
			m.recordFieldFocus = false
			return m, nil
		}
		if m.focus == FocusContent {
			m.focus = FocusTables
			m.viewMode = ViewSchema
			return m, nil
		}
		return m, nil
	case keyMatches(keyRuntimeFilter, key):
		return m.startFilterPopup()
	case keyMatches(keyRuntimeSort, key):
		return m.startSortPopup()
	case keyMatches(keyRuntimeRecordDetail, key):
		return m.openRecordDetail()
	case keyMatches(keyRuntimeSave, key):
		return m.requestSaveChanges()
	case keyMatches(keyRuntimeInsert, key):
		return m.addPendingInsert()
	case keyMatches(keyRuntimeDelete, key):
		return m.toggleDeleteSelection()
	case keyMatches(keyRuntimeUndo, key):
		return m.undoStagedAction()
	case keyMatches(keyRuntimeRedo, key):
		return m.redoStagedAction()
	case keyMatches(keyRuntimeToggleAutoFields, key):
		return m.toggleInsertAutoFields()
	case keyMatches(keyRuntimeMoveDown, key):
		return m.moveDown()
	case keyMatches(keyRuntimeMoveUp, key):
		return m.moveUp()
	case keyMatches(keyRuntimeMoveLeft, key):
		return m.moveLeft()
	case keyMatches(keyRuntimeMoveRight, key):
		return m.moveRight()
	case keyMatches(keyRuntimePageDown, key):
		return m.pageDown()
	case keyMatches(keyRuntimePageUp, key):
		return m.pageUp()
	default:
		return m, nil
	}
}
