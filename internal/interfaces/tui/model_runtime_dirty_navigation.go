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

func (m *Model) confirmOptionsFromDirtyPrompt(prompt usecase.DirtyDecisionPrompt, flow dirtyConfirmFlow) []confirmOption {
	options := make([]confirmOption, 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, confirmOption{
			label:  option.Label,
			action: mapDirtyDecisionToConfirmAction(option.ID, flow),
		})
	}
	return options
}

func mapDirtyDecisionToConfirmAction(decisionID string, flow dirtyConfirmFlow) confirmAction {
	switch flow {
	case dirtyConfirmFlowTableSwitch:
		switch decisionID {
		case usecase.DirtyDecisionDiscard:
			return confirmDiscardTable
		case usecase.DirtyDecisionCancel:
			return confirmCancelTableSwitch
		default:
			return confirmCancelTableSwitch
		}
	default:
		switch flow {
		case dirtyConfirmFlowConfig:
			switch decisionID {
			case usecase.DirtyDecisionSave:
				return confirmConfigSaveAndOpen
			case usecase.DirtyDecisionDiscard:
				return confirmConfigDiscardAndOpen
			case usecase.DirtyDecisionCancel:
				return confirmConfigCancel
			default:
				return confirmConfigCancel
			}
		case dirtyConfirmFlowQuit:
			switch decisionID {
			case usecase.DirtyDecisionDiscard:
				return confirmDiscardQuit
			case usecase.DirtyDecisionCancel:
				return confirmCancelQuit
			default:
				return confirmCancelQuit
			}
		default:
			return confirmConfigCancel
		}
	}
}
