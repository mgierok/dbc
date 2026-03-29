package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) confirmOptionsFromNavigationPrompt(prompt usecase.RuntimeNavigationDecisionPrompt) []confirmOption {
	options := make([]confirmOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, confirmOption{
			label:      option.Label,
			decisionID: option.ID,
		})
	}
	return options
}

func (m *Model) applyRuntimeNavigationPlan(plan usecase.RuntimeNavigationPlan, pendingCommandInput string) (tea.Model, tea.Cmd) {
	if plan.Prompt != nil && plan.Pending != nil {
		m.ui.pendingNavigation = clonePendingRuntimeNavigation(plan.Pending)
		m.ui.pendingCommandInput = pendingCommandInput
		if pendingCommandInput != "" {
			m.overlay.commandInput = commandInput{}
		}
		m.openModalConfirmPopupWithOptions(
			plan.Prompt.Title,
			plan.Prompt.Message,
			m.confirmOptionsFromNavigationPrompt(*plan.Prompt),
			0,
		)
		return m, nil
	}

	if pendingCommandInput != "" {
		m.overlay.commandInput = commandInput{}
	}
	return m.executeRuntimeNavigationNextAction(plan.NextAction)
}

func clonePendingRuntimeNavigation(pending *usecase.PendingRuntimeNavigation) *usecase.PendingRuntimeNavigation {
	if pending == nil {
		return nil
	}
	cloned := *pending
	return &cloned
}
