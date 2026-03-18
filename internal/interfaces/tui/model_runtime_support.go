package tui

import (
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) pageSize() int {
	height := m.contentHeight()
	if height < 4 {
		return 1
	}
	if m.read.focus == FocusContent && m.read.viewMode == ViewRecords {
		return height - 2
	}
	return height - 1
}

func (m *Model) effectiveRecordLimit() int {
	if m.runtimeSession == nil {
		return defaultRecordPageLimit
	}
	return m.runtimeSession.effectiveRecordsPageLimit()
}

func (m *Model) contentHeight() int {
	if m.ui.height <= 0 {
		return 16
	}
	if m.ui.height <= 6 {
		return 1
	}
	return m.ui.height - 5
}

func (m *Model) contentSelection() int {
	switch m.read.viewMode {
	case ViewSchema:
		return m.read.schemaIndex
	case ViewRecords:
		return m.read.recordSelection
	default:
		return 0
	}
}

func (m *Model) contentMaxIndex() int {
	switch m.read.viewMode {
	case ViewSchema:
		return len(m.read.schema.Columns) - 1
	case ViewRecords:
		return m.totalRecordRows() - 1
	default:
		return 0
	}
}

func (m *Model) currentTableName() string {
	if len(m.read.tables) == 0 {
		return ""
	}
	if m.read.selectedTable < 0 || m.read.selectedTable >= len(m.read.tables) {
		return ""
	}
	return m.read.tables[m.read.selectedTable].Name
}

func (m *Model) commandInputValueWithCaret() string {
	if !m.overlay.commandInput.active {
		return ""
	}
	cursor := clamp(m.overlay.commandInput.cursor, 0, len(m.overlay.commandInput.value))
	return m.overlay.commandInput.value[:cursor] + "|" + m.overlay.commandInput.value[cursor:]
}

func (m *Model) ShouldOpenConfigSelector() bool {
	return m.ui.openConfigSelector
}

func optionIndex(options []string, value string) int {
	for i, option := range options {
		if strings.EqualFold(option, value) {
			return i
		}
	}
	return 0
}

func sortDirections() []dto.SortDirection {
	return []dto.SortDirection{dto.SortDirectionAsc, dto.SortDirectionDesc}
}

func containsInt(values []int, target int) bool {
	return indexOfInt(values, target) >= 0
}

func indexOfInt(values []int, target int) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func insertAtCursor(value, insert string, cursor int) (string, int) {
	if insert == "" {
		return value, cursor
	}
	cursor = clamp(cursor, 0, len(value))
	updated := value[:cursor] + insert + value[cursor:]
	return updated, cursor + len(insert)
}

func deleteAtCursor(value string, cursor int) (string, int) {
	if value == "" || cursor <= 0 {
		return value, 0
	}
	cursor = clamp(cursor, 0, len(value))
	updated := value[:cursor-1] + value[cursor:]
	return updated, cursor - 1
}
