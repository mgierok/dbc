package tui

import (
	"github.com/mgierok/dbc/internal/application/usecase"
)

func (m *Model) dirtyNavigationPolicyUseCase() *usecase.DirtyNavigationPolicy {
	if m.dirtyNavPolicy != nil {
		return m.dirtyNavPolicy
	}
	return usecase.NewDirtyNavigationPolicy()
}

func (m *Model) confirmOptionsFromDirtyPrompt(prompt usecase.DirtyDecisionPrompt) []confirmOption {
	options := make([]confirmOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, confirmOption{
			label:  option.Label,
			action: mapDirtyDecisionToConfirmAction(option.ID),
		})
	}
	return options
}

func mapDirtyDecisionToConfirmAction(decisionID string) confirmAction {
	switch decisionID {
	case usecase.DirtyDecisionSave:
		return confirmLeaveSave
	case usecase.DirtyDecisionDiscard:
		return confirmLeaveDiscard
	case usecase.DirtyDecisionCancel:
		return confirmLeaveCancel
	default:
		return confirmLeaveCancel
	}
}
