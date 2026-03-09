package selector

func (m *databaseSelectorModel) moveSelection(delta int) {
	if len(m.options) == 0 {
		return
	}
	m.browse.selected = clamp(m.browse.selected+delta, 0, len(m.options)-1)
}

func (m *databaseSelectorModel) selectTop() {
	if len(m.options) == 0 {
		return
	}
	m.browse.selected = 0
}

func (m *databaseSelectorModel) selectBottom() {
	if len(m.options) == 0 {
		return
	}
	m.browse.selected = len(m.options) - 1
}

func (m *databaseSelectorModel) page(delta int) {
	page := m.listHeight(m.height)
	if page < 1 {
		page = 1
	}
	m.moveSelection(delta * page)
}

func (m *databaseSelectorModel) openAddForm() {
	m.mode = selectorModeAdd
	m.form = selectorForm{
		editIndex:   -1,
		activeField: selectorInputName,
	}
}

func (m *databaseSelectorModel) openEditForm() {
	if len(m.options) == 0 {
		m.browse.statusMessage = "No database selected to edit"
		return
	}
	selected := m.options[m.browse.selected]
	if !selected.isConfigBacked() {
		m.browse.statusMessage = "CLI session entry cannot be edited"
		return
	}
	m.mode = selectorModeEdit
	m.form = selectorForm{
		editIndex:   selected.managerIndex,
		activeField: selectorInputName,
		nameValue:   selected.Name,
		pathValue:   selected.ConnString,
	}
}

func (m *databaseSelectorModel) openDeleteConfirmation() {
	if len(m.options) == 0 {
		m.browse.statusMessage = "No database selected to delete"
		return
	}
	selected := m.options[m.browse.selected]
	if !selected.isConfigBacked() {
		m.browse.statusMessage = "CLI session entry cannot be deleted"
		return
	}
	m.mode = selectorModeConfirmDelete
	m.confirmDelete = selectorDeleteConfirm{
		active:       true,
		optionIndex:  m.browse.selected,
		managerIndex: selected.managerIndex,
	}
}
