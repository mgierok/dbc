package tui

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
	lines := m.boxLines(listHeight, width)
	boxHeight := len(lines)
	if boxHeight > height {
		boxHeight = height
		lines = lines[:boxHeight]
	}

	leftPad := 0
	boxWidth := textWidth(lines[0])
	if width > boxWidth {
		leftPad = (width - boxWidth) / 2
	}
	topPad := 0
	if height > boxHeight {
		topPad = (height - boxHeight) / 2
	}

	full := make([]string, 0, height)
	for i := 0; i < topPad; i++ {
		full = append(full, strings.Repeat(" ", width))
	}
	for _, line := range lines {
		full = append(full, padRight(strings.Repeat(" ", leftPad)+line, width))
	}
	for len(full) < height {
		full = append(full, strings.Repeat(" ", width))
	}
	return strings.Join(full, "\n")
}

func (m *databaseSelectorModel) renderHelpPopup(totalWidth, totalHeight int) []string {
	return renderStandardizedPopup(totalWidth, totalHeight, standardizedPopupSpec{
		title:               "Context Help: Config",
		summary:             runtimeHelpPopupSummaryLine(),
		rows:                m.helpPopupContentLines(),
		selected:            -1,
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
	pathLine := m.styles.summary("Config: " + configPath)

	bodyLines := m.mainContentLines(listHeight)
	hintLine := m.styles.muted(runtimeStatusContextHelpHint())

	contentAreaWidth := textWidth(title)
	if textWidth(pathLine) > contentAreaWidth {
		contentAreaWidth = textWidth(pathLine)
	}
	if textWidth(hintLine) > contentAreaWidth {
		contentAreaWidth = textWidth(hintLine)
	}
	for _, line := range bodyLines {
		if textWidth(line) > contentAreaWidth {
			contentAreaWidth = textWidth(line)
		}
	}
	if contentAreaWidth < 1 {
		contentAreaWidth = 1
	}

	contentInnerWidth := contentAreaWidth + (popupContentSidePadding * 2)
	maxInner := totalWidth - 2
	if maxInner < 1 {
		maxInner = 1
	}
	if contentInnerWidth > maxInner {
		contentInnerWidth = maxInner
	}

	leftPadding := popupContentSidePadding
	rightPadding := popupContentSidePadding
	if contentInnerWidth <= (popupContentSidePadding * 2) {
		leftPadding = 0
		rightPadding = 0
	}
	contentWidth := contentInnerWidth - leftPadding - rightPadding
	if contentWidth < 0 {
		contentWidth = 0
	}

	buildContentLine := func(text string) string {
		content := strings.Repeat(" ", leftPadding) + padRight(text, contentWidth) + strings.Repeat(" ", rightPadding)
		content = padRight(content, contentInnerWidth)
		return frameVertical + content + frameVertical
	}

	topBorder := renderTitledTopBorder(m.styles.title(title), contentInnerWidth)
	bottomBorder := frameBottomLeft + strings.Repeat(frameHorizontal, contentInnerWidth) + frameBottomRight
	sectionDivider := frameJoinLeft + strings.Repeat(frameHorizontal, contentInnerWidth) + frameJoinRight
	lines := []string{
		topBorder,
		buildContentLine(pathLine),
		sectionDivider,
	}
	for _, line := range bodyLines {
		framedLine := buildContentLine(line)
		if strings.HasPrefix(line, selectionSelectedPrefix()) {
			framedLine = m.styles.selected(framedLine)
		}
		lines = append(lines, framedLine)
	}

	minHeight := popupMinHeight(m.height)
	for len(lines)+2 < minHeight {
		lines = append(lines, buildContentLine(""))
	}
	lines = append(lines, buildContentLine(renderStatusWithRightHint("", hintLine, contentWidth)))
	lines = append(lines, bottomBorder)
	return lines
}

func (m *databaseSelectorModel) optionLines() []string {
	items := make([]string, len(m.options))
	for i, option := range m.options {
		items[i] = option.marker() + " " + option.Name + frameSegmentSeparator + option.ConnString
	}
	return items
}

func (m *databaseSelectorModel) mainContentLines(listHeight int) []string {
	switch m.mode {
	case selectorModeAdd, selectorModeEdit:
		return m.formContentLines()
	case selectorModeConfirmDelete:
		return m.deleteConfirmationContentLines()
	default:
		return m.browseContentLines(listHeight)
	}
}

func (m *databaseSelectorModel) browseContentLines(listHeight int) []string {
	items := m.optionLines()
	lines := make([]string, 0, maxInt(1, len(items)))
	if len(items) == 0 {
		lines = append(lines, "No databases configured.")
	} else {
		start := scrollStart(m.selected, listHeight, len(items))
		end := minInt(len(items), start+listHeight)
		for i := start; i < end; i++ {
			prefix := selectionUnselectedPrefix()
			if i == m.selected {
				prefix = selectionSelectedPrefix()
			}
			lines = append(lines, prefix+items[i])
		}
	}
	if strings.TrimSpace(m.statusMessage) != "" {
		lines = append(lines, "Status: "+m.styleStatusMessage())
	}
	return lines
}

func (m *databaseSelectorModel) formContentLines() []string {
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

	lines := []string{
		title,
		"",
		m.styleFormLine(namePrefix, "Name", nameValue, m.form.activeField == selectorInputName),
		m.styleFormLine(pathPrefix, "Path", pathValue, m.form.activeField == selectorInputPath),
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		lines = append(lines, "", m.styles.error("Error: "+m.form.errorMessage))
	}
	return lines
}

func (m *databaseSelectorModel) deleteConfirmationContentLines() []string {
	if m.mode != selectorModeConfirmDelete {
		return nil
	}
	if m.confirmDelete.optionIndex < 0 || m.confirmDelete.optionIndex >= len(m.options) {
		return []string{
			"Cannot delete: invalid selection.",
			"Press Esc to return.",
		}
	}
	selected := m.options[m.confirmDelete.optionIndex]
	return []string{
		"Delete database entry?",
		"",
		selected.Name + frameSegmentSeparator + selected.ConnString,
	}
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
		m.styleFormLine(namePrefix, "Name", nameValue, m.form.activeField == selectorInputName),
		m.styleFormLine(pathPrefix, "Path", pathValue, m.form.activeField == selectorInputPath),
		"",
		selectorFormSwitchLine(),
		selectorFormSubmitLine(escLabel),
	}
	if strings.TrimSpace(m.form.errorMessage) != "" {
		lines = append(lines, m.styles.error("Error: "+m.form.errorMessage))
	}
	return lines
}

func (m *databaseSelectorModel) styleFormLine(prefix, label, value string, active bool) string {
	line := prefix + label + ": " + value
	return line
}

func (m *databaseSelectorModel) styleStatusMessage() string {
	if isErrorLikeMessage(m.statusMessage) {
		return m.styles.error(m.statusMessage)
	}
	return m.statusMessage
}
