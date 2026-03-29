package usecase

type RuntimeNavigationActionKind int

const (
	RuntimeNavigationActionNone RuntimeNavigationActionKind = iota
	RuntimeNavigationActionSwitchTable
	RuntimeNavigationActionOpenDatabase
	RuntimeNavigationActionQuitRuntime
)

type RuntimeNavigationAction struct {
	Kind            RuntimeNavigationActionKind
	TargetTableName string
	DatabaseTarget  RuntimeDatabaseTarget
}

type PendingRuntimeNavigation struct {
	Action RuntimeNavigationAction
}

type RuntimeNavigationDecisionPrompt struct {
	Title   string
	Message string
	Options []DirtyDecisionOption
}

type RuntimeNavigationNextActionKind int

const (
	RuntimeNavigationNextActionNone RuntimeNavigationNextActionKind = iota
	RuntimeNavigationNextActionStayInRuntime
	RuntimeNavigationNextActionStartSave
	RuntimeNavigationNextActionSwitchTable
	RuntimeNavigationNextActionOpenDatabase
	RuntimeNavigationNextActionQuitRuntime
)

type RuntimeNavigationNextAction struct {
	Kind            RuntimeNavigationNextActionKind
	TargetTableName string
	DatabaseTarget  RuntimeDatabaseTarget
	ClearDirtyState bool
}

type RuntimeNavigationPlan struct {
	Prompt     *RuntimeNavigationDecisionPrompt
	Pending    *PendingRuntimeNavigation
	NextAction RuntimeNavigationNextAction
}

type RuntimeNavigationDecisionResult struct {
	Pending    *PendingRuntimeNavigation
	NextAction RuntimeNavigationNextAction
}

type RuntimeNavigationWorkflow struct {
	dirtyPolicy *DirtyNavigationPolicy
}

func NewRuntimeNavigationWorkflow() *RuntimeNavigationWorkflow {
	return &RuntimeNavigationWorkflow{
		dirtyPolicy: NewDirtyNavigationPolicy(),
	}
}

func (w *RuntimeNavigationWorkflow) PlanTableSwitch(targetTableName string, hasDirty bool, changeCount int) RuntimeNavigationPlan {
	action := RuntimeNavigationAction{
		Kind:            RuntimeNavigationActionSwitchTable,
		TargetTableName: targetTableName,
	}
	if !hasDirty {
		return RuntimeNavigationPlan{
			NextAction: w.executeAction(action, false),
		}
	}

	prompt := w.dirtyNavigationPolicy().BuildTableSwitchPrompt(changeCount)
	return RuntimeNavigationPlan{
		Prompt:  runtimeNavigationDecisionPrompt(prompt),
		Pending: &PendingRuntimeNavigation{Action: action},
	}
}

func (w *RuntimeNavigationWorkflow) PlanDatabaseTransition(target RuntimeDatabaseTarget, hasDirty bool, changeCount int) RuntimeNavigationPlan {
	action := RuntimeNavigationAction{
		Kind:           RuntimeNavigationActionOpenDatabase,
		DatabaseTarget: target,
	}
	if !hasDirty {
		return RuntimeNavigationPlan{
			NextAction: w.executeAction(action, false),
		}
	}

	prompt := w.dirtyNavigationPolicy().BuildDatabaseOpenPrompt(changeCount)
	if target.TransitionKind == RuntimeDatabaseTransitionReloadCurrent {
		prompt = w.dirtyNavigationPolicy().BuildDatabaseReloadPrompt(changeCount)
	}
	return RuntimeNavigationPlan{
		Prompt:  runtimeNavigationDecisionPrompt(prompt),
		Pending: &PendingRuntimeNavigation{Action: action},
	}
}

func (w *RuntimeNavigationWorkflow) PlanQuit(hasDirty bool, changeCount int) RuntimeNavigationPlan {
	action := RuntimeNavigationAction{Kind: RuntimeNavigationActionQuitRuntime}
	if !hasDirty {
		return RuntimeNavigationPlan{
			NextAction: w.executeAction(action, false),
		}
	}

	prompt := w.dirtyNavigationPolicy().BuildQuitPrompt(changeCount)
	return RuntimeNavigationPlan{
		Prompt:  runtimeNavigationDecisionPrompt(prompt),
		Pending: &PendingRuntimeNavigation{Action: action},
	}
}

func (w *RuntimeNavigationWorkflow) ResolveDecision(decisionID string, pending PendingRuntimeNavigation) RuntimeNavigationDecisionResult {
	switch decisionID {
	case DirtyDecisionSave:
		if pending.Action.Kind == RuntimeNavigationActionOpenDatabase {
			cloned := pending
			return RuntimeNavigationDecisionResult{
				Pending: &cloned,
				NextAction: RuntimeNavigationNextAction{
					Kind: RuntimeNavigationNextActionStartSave,
				},
			}
		}
	case DirtyDecisionDiscard:
		return RuntimeNavigationDecisionResult{
			NextAction: w.executeAction(pending.Action, true),
		}
	case DirtyDecisionCancel:
		return RuntimeNavigationDecisionResult{
			NextAction: RuntimeNavigationNextAction{
				Kind: RuntimeNavigationNextActionStayInRuntime,
			},
		}
	}

	return RuntimeNavigationDecisionResult{
		NextAction: RuntimeNavigationNextAction{
			Kind: RuntimeNavigationNextActionStayInRuntime,
		},
	}
}

func (w *RuntimeNavigationWorkflow) ResolveSaveResult(pending *PendingRuntimeNavigation, err error) RuntimeNavigationDecisionResult {
	if pending == nil {
		return RuntimeNavigationDecisionResult{
			NextAction: RuntimeNavigationNextAction{Kind: RuntimeNavigationNextActionStayInRuntime},
		}
	}
	if err != nil {
		cloned := *pending
		return RuntimeNavigationDecisionResult{
			Pending:    &cloned,
			NextAction: RuntimeNavigationNextAction{Kind: RuntimeNavigationNextActionStayInRuntime},
		}
	}
	return RuntimeNavigationDecisionResult{
		NextAction: w.executeAction(pending.Action, false),
	}
}

func (w *RuntimeNavigationWorkflow) executeAction(action RuntimeNavigationAction, clearDirtyState bool) RuntimeNavigationNextAction {
	switch action.Kind {
	case RuntimeNavigationActionSwitchTable:
		return RuntimeNavigationNextAction{
			Kind:            RuntimeNavigationNextActionSwitchTable,
			TargetTableName: action.TargetTableName,
			ClearDirtyState: clearDirtyState,
		}
	case RuntimeNavigationActionOpenDatabase:
		return RuntimeNavigationNextAction{
			Kind:            RuntimeNavigationNextActionOpenDatabase,
			DatabaseTarget:  action.DatabaseTarget,
			ClearDirtyState: clearDirtyState,
		}
	case RuntimeNavigationActionQuitRuntime:
		return RuntimeNavigationNextAction{
			Kind:            RuntimeNavigationNextActionQuitRuntime,
			ClearDirtyState: clearDirtyState,
		}
	default:
		return RuntimeNavigationNextAction{Kind: RuntimeNavigationNextActionStayInRuntime}
	}
}

func (w *RuntimeNavigationWorkflow) dirtyNavigationPolicy() *DirtyNavigationPolicy {
	if w != nil && w.dirtyPolicy != nil {
		return w.dirtyPolicy
	}
	return NewDirtyNavigationPolicy()
}

func runtimeNavigationDecisionPrompt(prompt DirtyDecisionPrompt) *RuntimeNavigationDecisionPrompt {
	return &RuntimeNavigationDecisionPrompt{
		Title:   prompt.Title,
		Message: prompt.Message,
		Options: append([]DirtyDecisionOption(nil), prompt.Options...),
	}
}
