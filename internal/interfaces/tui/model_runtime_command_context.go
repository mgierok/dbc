package tui

// runtime command entry is available only in non-blocking runtime contexts that
// own the active status bar workflow.
func (m *Model) commandInputSupportedInCurrentContext() bool {
	return m.nonBlockingRuntimeCommandContextActive()
}

func (m *Model) saveSupportedInCurrentContext() bool {
	return m.nonBlockingRuntimeCommandContextActive() && m.hasDirtyEdits()
}

func (m *Model) nonBlockingRuntimeCommandContextActive() bool {
	switch {
	case m.overlay.helpPopup.active:
		return false
	case m.overlay.filterPopup.active:
		return false
	case m.overlay.sortPopup.active:
		return false
	case m.overlay.editPopup.active:
		return false
	case m.overlay.confirmPopup.active:
		return false
	case m.overlay.recordDetail.active:
		return true
	case m.read.focus == FocusTables:
		return true
	case m.read.focus == FocusContent && m.read.viewMode == ViewSchema:
		return true
	case m.read.focus == FocusContent && m.read.viewMode == ViewRecords:
		return true
	default:
		return false
	}
}
