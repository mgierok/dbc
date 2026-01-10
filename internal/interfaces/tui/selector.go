package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	ErrNoDatabases                 = errors.New("no databases configured")
	ErrDatabaseSelectionCanceled   = errors.New("database selection canceled")
	ErrDatabaseSelectionUnfinished = errors.New("database selection not confirmed")
)

type DatabaseOption struct {
	Name       string
	ConnString string
}

type databaseSelectorModel struct {
	options  []DatabaseOption
	selected int
	width    int
	height   int
	chosen   bool
	canceled bool
}

func SelectDatabase(options []DatabaseOption) (DatabaseOption, error) {
	if len(options) == 0 {
		return DatabaseOption{}, ErrNoDatabases
	}
	model := newDatabaseSelectorModel(options)
	program := tea.NewProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return DatabaseOption{}, err
	}
	selector, ok := final.(*databaseSelectorModel)
	if !ok {
		return DatabaseOption{}, errors.New("unexpected selector state")
	}
	if selector.canceled {
		return DatabaseOption{}, ErrDatabaseSelectionCanceled
	}
	if !selector.chosen {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	if selector.selected < 0 || selector.selected >= len(selector.options) {
		return DatabaseOption{}, ErrDatabaseSelectionUnfinished
	}
	return selector.options[selector.selected], nil
}

func newDatabaseSelectorModel(options []DatabaseOption) *databaseSelectorModel {
	return &databaseSelectorModel{
		options: options,
	}
}

func (m *databaseSelectorModel) Init() tea.Cmd {
	return nil
}

func (m *databaseSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			if len(m.options) == 0 {
				return m, nil
			}
			m.chosen = true
			return m, tea.Quit
		case "j", "down":
			m.moveSelection(1)
			return m, nil
		case "k", "up":
			m.moveSelection(-1)
			return m, nil
		case "g", "home":
			m.selectTop()
			return m, nil
		case "G", "end":
			m.selectBottom()
			return m, nil
		case "ctrl+f", "pgdown":
			m.page(1)
			return m, nil
		case "ctrl+b", "pgup":
			m.page(-1)
			return m, nil
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m *databaseSelectorModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	listHeight := m.listHeight(height)
	lines := m.boxLines(listHeight, width)
	boxHeight := len(lines)
	if boxHeight > height {
		boxHeight = height
		lines = lines[:boxHeight]
	}

	leftPad := 0
	boxWidth := len(lines[0])
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

func (m *databaseSelectorModel) listHeight(totalHeight int) int {
	if totalHeight <= 0 {
		totalHeight = 24
	}
	maxListHeight := totalHeight - 4
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
	items := m.optionLines()
	contentWidth := len(title)
	for _, item := range items {
		itemWidth := len(item) + 2
		if itemWidth > contentWidth {
			contentWidth = itemWidth
		}
	}
	if contentWidth < 1 {
		contentWidth = 1
	}
	maxInner := totalWidth - 2
	if maxInner < 1 {
		maxInner = 1
	}
	if contentWidth > maxInner {
		contentWidth = maxInner
	}

	border := "+" + strings.Repeat("-", contentWidth) + "+"
	lines := []string{
		border,
		"|" + padRight(title, contentWidth) + "|",
		"|" + strings.Repeat("-", contentWidth) + "|",
	}

	if len(items) == 0 {
		lines = append(lines, "|"+padRight("No databases configured.", contentWidth)+"|")
		lines = append(lines, border)
		return lines
	}

	start := scrollStart(m.selected, listHeight, len(items))
	end := minInt(len(items), start+listHeight)
	for i := start; i < end; i++ {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		lines = append(lines, "|"+padRight(prefix+items[i], contentWidth)+"|")
	}
	lines = append(lines, border)
	return lines
}

func (m *databaseSelectorModel) optionLines() []string {
	items := make([]string, len(m.options))
	for i, option := range m.options {
		items[i] = fmt.Sprintf("%s | %s", option.Name, option.ConnString)
	}
	return items
}

func (m *databaseSelectorModel) moveSelection(delta int) {
	if len(m.options) == 0 {
		return
	}
	m.selected = clamp(m.selected+delta, 0, len(m.options)-1)
}

func (m *databaseSelectorModel) selectTop() {
	if len(m.options) == 0 {
		return
	}
	m.selected = 0
}

func (m *databaseSelectorModel) selectBottom() {
	if len(m.options) == 0 {
		return
	}
	m.selected = len(m.options) - 1
}

func (m *databaseSelectorModel) page(delta int) {
	page := m.listHeight(m.height)
	if page < 1 {
		page = 1
	}
	m.moveSelection(delta * page)
}
