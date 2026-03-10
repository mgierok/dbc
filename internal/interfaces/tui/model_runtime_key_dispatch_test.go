package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKey_CtrlWPanelShortcutsAreUnsupported(t *testing.T) {
	tests := []struct {
		name       string
		startFocus PanelFocus
		nextKey    tea.KeyMsg
	}{
		{
			name:       "ctrl+w h does not switch to tables",
			startFocus: FocusContent,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
		},
		{
			name:       "ctrl+w l does not switch to content",
			startFocus: FocusTables,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
		},
		{
			name:       "ctrl+w w does not toggle panel",
			startFocus: FocusTables,
			nextKey:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			model := &Model{
				viewMode: ViewRecords,
				focus:    tc.startFocus,
			}

			// Act
			model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlW})
			model.handleKey(tc.nextKey)

			// Assert
			if model.focus != tc.startFocus {
				t.Fatalf("expected focus to stay %v, got %v", tc.startFocus, model.focus)
			}
		})
	}
}
