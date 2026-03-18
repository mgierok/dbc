package selector

import (
	"strings"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (m *databaseSelectorModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	return primitives.CenterBoxLines(m.popupLines(width, height), width, height)
}

func (m *databaseSelectorModel) popupLines(totalWidth, totalHeight int) []string {
	if m.helpPopup.active {
		return m.renderHelpPopup(totalWidth, totalHeight)
	}
	listHeight := m.listHeight(totalHeight)
	return m.boxLines(listHeight, totalWidth, totalHeight)
}

func (m *databaseSelectorModel) renderHelpPopup(totalWidth, totalHeight int) []string {
	return primitives.RenderStandardizedPopup(totalWidth, totalHeight, primitives.StandardizedPopupSpec{
		Title:               "Context Help: Config",
		Summary:             primitives.RuntimeHelpPopupSummaryLine(),
		Rows:                primitives.PopupTextRows(m.helpPopupContentLines()),
		ScrollOffset:        m.helpPopup.scrollOffset,
		VisibleRows:         m.helpPopupVisibleLines(),
		ShowScrollIndicator: true,
		DefaultWidth:        60,
		MinWidth:            20,
		MaxWidth:            70,
		Styles:              m.styles,
	})
}

func (m *databaseSelectorModel) helpPopupContentLines() []string {
	switch m.helpPopup.context {
	case selectorModeAdd, selectorModeEdit:
		escLabel := "Esc cancel"
		if m.requiresFirstEntry && len(m.options) == 0 && m.helpPopup.context == selectorModeAdd {
			escLabel = "Esc " + m.firstSetupEscActionLabel()
		}
		return []string{
			primitives.SelectorFormSwitchLine(),
			primitives.SelectorFormSubmitLine(escLabel),
		}
	case selectorModeConfirmDelete:
		return []string{primitives.SelectorDeleteConfirmationLine()}
	case selectorModeBrowse:
		if m.requiresFirstEntry {
			return primitives.SelectorContextLinesBrowseFirstSetup(m.firstSetupEscActionLabel())
		}
		return primitives.SelectorContextLinesBrowseDefault(m.browseEscActionLabel())
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

func (m *databaseSelectorModel) boxLines(listHeight, totalWidth int, totalHeight ...int) []string {
	title := "Select database"
	configPath := m.activeConfigPath
	if strings.TrimSpace(configPath) == "" {
		configPath = "unavailable"
	}
	resolvedHeight := m.height
	if len(totalHeight) > 0 {
		resolvedHeight = totalHeight[0]
	}
	return primitives.RenderStandardizedPopup(totalWidth, resolvedHeight, primitives.StandardizedPopupSpec{
		Title:     title,
		Summary:   "Config: " + configPath,
		Rows:      m.mainContentRows(listHeight),
		Footer:    primitives.StandardizedPopupFooter{Right: m.styles.Muted(primitives.RuntimeStatusContextHelpHint())},
		WidthMode: primitives.PopupWidthContent,
		Styles:    m.styles,
	})
}

func (m *databaseSelectorModel) optionLines() []string {
	items := make([]string, len(m.options))
	for i, option := range m.options {
		items[i] = option.marker() + " " + option.Name + primitives.FrameSegmentSeparator + option.ConnString
	}
	return items
}

func (m *databaseSelectorModel) mainContentRows(listHeight int) []primitives.StandardizedPopupRow {
	switch m.mode {
	case selectorModeAdd, selectorModeEdit:
		return m.formContentRows()
	case selectorModeConfirmDelete:
		return m.deleteConfirmationContentRows()
	default:
		return m.browseContentRows(listHeight)
	}
}

func (m *databaseSelectorModel) browseContentRows(listHeight int) []primitives.StandardizedPopupRow {
	items := m.optionLines()
	rows := make([]primitives.StandardizedPopupRow, 0, primitives.MaxInt(1, len(items)))
	if len(items) == 0 {
		rows = append(rows, primitives.StandardizedPopupRow{Text: "No databases configured."})
	} else {
		start := primitives.ScrollStart(m.browse.selected, listHeight, len(items))
		end := primitives.MinInt(len(items), start+listHeight)
		for i := start; i < end; i++ {
			rows = append(rows, primitives.StandardizedPopupRow{
				Text:       items[i],
				Selectable: true,
				Selected:   i == m.browse.selected,
			})
		}
	}
	if strings.TrimSpace(m.browse.statusMessage) != "" {
		rows = append(rows, primitives.StandardizedPopupRow{Text: "Status: " + m.styleStatusMessage()})
	}
	return rows
}

func (m *databaseSelectorModel) formContentRows() []primitives.StandardizedPopupRow {
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

	rows := []primitives.StandardizedPopupRow{
		{Text: title},
		{Text: ""},
		{Text: selectorFormFieldLine("Name", nameValue), Selectable: true, Selected: m.form.activeField == selectorInputName},
		{Text: selectorFormFieldLine("Path", pathValue), Selectable: true, Selected: m.form.activeField == selectorInputPath},
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		rows = append(rows,
			primitives.StandardizedPopupRow{Text: ""},
			primitives.StandardizedPopupRow{Text: m.styles.Error("Error: " + m.form.errorMessage)},
		)
	}
	return rows
}

func (m *databaseSelectorModel) deleteConfirmationContentRows() []primitives.StandardizedPopupRow {
	if m.mode != selectorModeConfirmDelete {
		return nil
	}
	if m.confirmDelete.optionIndex < 0 || m.confirmDelete.optionIndex >= len(m.options) {
		return primitives.PopupTextRows([]string{
			"Cannot delete: invalid selection.",
			"Press Esc to return.",
		})
	}
	selected := m.options[m.confirmDelete.optionIndex]
	return primitives.PopupTextRows([]string{
		"Delete database entry?",
		"",
		selected.Name + primitives.FrameSegmentSeparator + selected.ConnString,
	})
}

func selectorFormFieldLine(label, value string) string {
	return label + ": " + value
}

func (m *databaseSelectorModel) styleStatusMessage() string {
	if primitives.IsErrorLikeMessage(m.browse.statusMessage) {
		return m.styles.Error(m.browse.statusMessage)
	}
	return m.browse.statusMessage
}
