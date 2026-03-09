package selector

import "strings"

func (m *databaseSelectorModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if m.helpPopup.active {
		return centerBoxLines(m.renderHelpPopup(width, height), width, height)
	}

	listHeight := m.listHeight(height)
	return centerBoxLines(m.boxLines(listHeight, width), width, height)
}

func (m *databaseSelectorModel) renderHelpPopup(totalWidth, totalHeight int) []string {
	return renderStandardizedPopup(totalWidth, totalHeight, standardizedPopupSpec{
		title:               "Context Help: Config",
		summary:             runtimeHelpPopupSummaryLine(),
		rows:                popupTextRows(m.helpPopupContentLines()),
		scrollOffset:        m.helpPopup.scrollOffset,
		visibleRows:         m.helpPopupVisibleLines(),
		showScrollIndicator: true,
		defaultWidth:        60,
		minWidth:            20,
		maxWidth:            70,
		styles:              m.styles,
	})
}

func (m *databaseSelectorModel) helpPopupContentLines() []string {
	switch m.helpPopup.context {
	case selectorModeAdd, selectorModeEdit:
		escLabel := "Esc cancel"
		if m.requiresFirstEntry && len(m.options) == 0 && m.helpPopup.context == selectorModeAdd {
			escLabel = "Esc exit app"
		}
		return []string{
			selectorFormSwitchLine(),
			selectorFormSubmitLine(escLabel),
		}
	case selectorModeConfirmDelete:
		return []string{selectorDeleteConfirmationLine()}
	case selectorModeBrowse:
		if m.requiresFirstEntry {
			return selectorContextLinesBrowseFirstSetup()
		}
		return selectorContextLinesBrowseDefault()
	default:
		return []string{"No shortcuts available."}
	}
}

func (m *databaseSelectorModel) listHeight(totalHeight int) int {
	if totalHeight <= 0 {
		totalHeight = 24
	}
	maxListHeight := totalHeight - 8
	if maxListHeight < 1 {
		maxListHeight = 1
	}
	if len(m.options) < maxListHeight {
		return len(m.options)
	}
	return maxListHeight
}

func (m *databaseSelectorModel) boxLines(listHeight, totalWidth int) []string {
	title := "Select database"
	configPath := m.activeConfigPath
	if strings.TrimSpace(configPath) == "" {
		configPath = "unavailable"
	}
	return renderStandardizedPopup(totalWidth, m.height, standardizedPopupSpec{
		title:     title,
		summary:   "Config: " + configPath,
		rows:      m.mainContentRows(listHeight),
		footer:    standardizedPopupFooter{right: m.styles.muted(runtimeStatusContextHelpHint())},
		widthMode: popupWidthContent,
		styles:    m.styles,
	})
}

func (m *databaseSelectorModel) optionLines() []string {
	items := make([]string, len(m.options))
	for i, option := range m.options {
		items[i] = option.marker() + " " + option.Name + frameSegmentSeparator + option.ConnString
	}
	return items
}

func (m *databaseSelectorModel) mainContentRows(listHeight int) []standardizedPopupRow {
	switch m.mode {
	case selectorModeAdd, selectorModeEdit:
		return m.formContentRows()
	case selectorModeConfirmDelete:
		return m.deleteConfirmationContentRows()
	default:
		return m.browseContentRows(listHeight)
	}
}

func (m *databaseSelectorModel) browseContentRows(listHeight int) []standardizedPopupRow {
	items := m.optionLines()
	rows := make([]standardizedPopupRow, 0, maxInt(1, len(items)))
	if len(items) == 0 {
		rows = append(rows, standardizedPopupRow{text: "No databases configured."})
	} else {
		start := scrollStart(m.browse.selected, listHeight, len(items))
		end := minInt(len(items), start+listHeight)
		for i := start; i < end; i++ {
			rows = append(rows, standardizedPopupRow{
				text:       items[i],
				selectable: true,
				selected:   i == m.browse.selected,
			})
		}
	}
	if strings.TrimSpace(m.browse.statusMessage) != "" {
		rows = append(rows, standardizedPopupRow{text: "Status: " + m.styleStatusMessage()})
	}
	return rows
}

func (m *databaseSelectorModel) formContentRows() []standardizedPopupRow {
	if m.mode != selectorModeAdd && m.mode != selectorModeEdit {
		return nil
	}
	title := "Add database"
	if m.mode == selectorModeEdit {
		title = "Edit database"
	}
	nameValue := m.form.nameValue
	pathValue := m.form.pathValue
	if m.form.activeField == selectorInputName {
		nameValue += "|"
	} else {
		pathValue += "|"
	}

	rows := []standardizedPopupRow{
		{text: title},
		{text: ""},
		{text: selectorFormFieldLine("Name", nameValue), selectable: true, selected: m.form.activeField == selectorInputName},
		{text: selectorFormFieldLine("Path", pathValue), selectable: true, selected: m.form.activeField == selectorInputPath},
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		rows = append(rows,
			standardizedPopupRow{text: ""},
			standardizedPopupRow{text: m.styles.error("Error: " + m.form.errorMessage)},
		)
	}
	return rows
}

func (m *databaseSelectorModel) deleteConfirmationContentRows() []standardizedPopupRow {
	if m.mode != selectorModeConfirmDelete {
		return nil
	}
	if m.confirmDelete.optionIndex < 0 || m.confirmDelete.optionIndex >= len(m.options) {
		return popupTextRows([]string{
			"Cannot delete: invalid selection.",
			"Press Esc to return.",
		})
	}
	selected := m.options[m.confirmDelete.optionIndex]
	return popupTextRows([]string{
		"Delete database entry?",
		"",
		selected.Name + frameSegmentSeparator + selected.ConnString,
	})
}

func (m *databaseSelectorModel) formLines() []string {
	if m.mode != selectorModeAdd && m.mode != selectorModeEdit {
		return nil
	}
	title := "Add database"
	if m.mode == selectorModeEdit {
		title = "Edit database"
	}
	namePrefix := selectionUnselectedPrefix()
	pathPrefix := selectionUnselectedPrefix()
	nameValue := m.form.nameValue
	pathValue := m.form.pathValue
	if m.form.activeField == selectorInputName {
		namePrefix = selectionSelectedPrefix()
		nameValue += "|"
	} else {
		pathPrefix = selectionSelectedPrefix()
		pathValue += "|"
	}

	escLabel := "Esc cancel"
	if m.requiresFirstEntry && len(m.options) == 0 && m.mode == selectorModeAdd {
		escLabel = "Esc exit app"
	}

	lines := []string{
		title,
		"",
		selectorFormFieldLineWithPrefix(namePrefix, "Name", nameValue),
		selectorFormFieldLineWithPrefix(pathPrefix, "Path", pathValue),
		"",
		selectorFormSwitchLine(),
		selectorFormSubmitLine(escLabel),
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		lines = append(lines, m.styles.error("Error: "+m.form.errorMessage))
	}
	return lines
}

func selectorFormFieldLineWithPrefix(prefix, label, value string) string {
	return prefix + selectorFormFieldLine(label, value)
}

func selectorFormFieldLine(label, value string) string {
	return label + ": " + value
}

func (m *databaseSelectorModel) styleStatusMessage() string {
	if isErrorLikeMessage(m.browse.statusMessage) {
		return m.styles.error(m.browse.statusMessage)
	}
	return m.browse.statusMessage
}
