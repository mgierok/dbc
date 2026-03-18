package selector

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *databaseSelectorModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case primitives.KeyMatches(primitives.KeySelectorFormEsc, key):
		if m.requiresFirstEntry && len(m.options) == 0 {
			m.requestClose()
			return m, nil
		}
		m.mode = selectorModeBrowse
		m.form = selectorForm{}
		m.browse.statusMessage = "Config update canceled"
		return m, nil
	case primitives.KeyMatches(primitives.KeySelectorFormSwitch, key):
		if m.form.activeField == selectorInputName {
			m.form.activeField = selectorInputPath
		} else {
			m.form.activeField = selectorInputName
		}
		return m, nil
	case primitives.KeyMatches(primitives.KeySelectorFormClear, key):
		m.setActiveFormValue("")
		m.form.errorMessage = ""
		return m, nil
	case primitives.KeyMatches(primitives.KeySelectorFormBackspace, key):
		value := m.activeFormValue()
		if value == "" {
			return m, nil
		}
		m.setActiveFormValue(value[:len(value)-1])
		m.form.errorMessage = ""
		return m, nil
	case primitives.KeyMatches(primitives.KeySelectorEnter, key):
		return m.submitForm()
	default:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		m.setActiveFormValue(m.activeFormValue() + string(msg.Runes))
		m.form.errorMessage = ""
		return m, nil
	}
}

func (m *databaseSelectorModel) submitForm() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.form.nameValue)
	path := strings.TrimSpace(m.form.pathValue)
	if name == "" {
		m.form.errorMessage = "Name is required"
		return m, nil
	}
	if path == "" {
		m.form.errorMessage = "Path is required"
		return m, nil
	}

	entry := dto.ConfigDatabase{
		Name: name,
		Path: path,
	}

	var err error
	switch m.mode {
	case selectorModeAdd:
		err = m.manager.Create(m.ctx, entry)
	case selectorModeEdit:
		err = m.manager.Update(m.ctx, m.form.editIndex, entry)
	default:
		return m, nil
	}
	if err != nil {
		m.form.errorMessage = err.Error()
		return m, nil
	}

	previousMode := m.mode
	previousEditIndex := m.form.editIndex
	m.mode = selectorModeBrowse
	m.form = selectorForm{}

	if err := m.refreshOptions(); err != nil {
		m.browse.statusMessage = "Config updated but refresh failed: " + err.Error()
		return m, nil
	}
	if previousMode == selectorModeAdd && len(m.options) > 0 {
		m.browse.selected = m.configOptionCount - 1
		if m.requiresFirstEntry {
			m.browse.statusMessage = "Database added. Press Enter to continue or a to add another"
			return m, nil
		}
		m.browse.statusMessage = "Database added"
		return m, nil
	}
	if previousMode == selectorModeEdit && m.configOptionCount > 0 {
		m.browse.selected = clamp(previousEditIndex, 0, m.configOptionCount-1)
		m.browse.statusMessage = "Database updated"
	}
	return m, nil
}

func (m *databaseSelectorModel) activeFormValue() string {
	if m.form.activeField == selectorInputPath {
		return m.form.pathValue
	}
	return m.form.nameValue
}

func (m *databaseSelectorModel) setActiveFormValue(value string) {
	if m.form.activeField == selectorInputPath {
		m.form.pathValue = value
		return
	}
	m.form.nameValue = value
}
