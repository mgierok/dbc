package usecase

import "fmt"

type RuntimeSaveIntent int

const (
	RuntimeSaveIntentSaveOnly RuntimeSaveIntent = iota + 1
	RuntimeSaveIntentSaveAndQuit
	RuntimeSaveIntentSaveForPendingTransition
)

type RuntimeSaveSuccessAction int

const (
	RuntimeSaveSuccessActionNone RuntimeSaveSuccessAction = iota
	RuntimeSaveSuccessActionStayInRuntime
	RuntimeSaveSuccessActionQuitRuntime
	RuntimeSaveSuccessActionRunPendingTransition
)

type RuntimeSaveResultNextAction int

const (
	RuntimeSaveResultNextActionNone RuntimeSaveResultNextAction = iota
	RuntimeSaveResultNextActionReloadRecords
	RuntimeSaveResultNextActionQuitRuntime
	RuntimeSaveResultNextActionRunPendingTransition
)

type RuntimeSaveRequestDecision struct {
	StartSave       bool
	ImmediateStatus string
	ImmediateExit   bool
	SuccessAction   RuntimeSaveSuccessAction
}

type RuntimeSaveResultDecision struct {
	ClearStaging           bool
	StatusMessage          string
	NextAction             RuntimeSaveResultNextAction
	ClearPendingSaveAction bool
}

type RuntimeSaveWorkflow struct{}

func NewRuntimeSaveWorkflow() *RuntimeSaveWorkflow {
	return &RuntimeSaveWorkflow{}
}

func (w *RuntimeSaveWorkflow) PlanRequest(intent RuntimeSaveIntent, hasDirty bool) RuntimeSaveRequestDecision {
	if hasDirty {
		return RuntimeSaveRequestDecision{StartSave: true}
	}

	switch intent {
	case RuntimeSaveIntentSaveOnly:
		return RuntimeSaveRequestDecision{ImmediateStatus: "No changes to save"}
	case RuntimeSaveIntentSaveAndQuit:
		return RuntimeSaveRequestDecision{ImmediateExit: true}
	default:
		return RuntimeSaveRequestDecision{}
	}
}

func (w *RuntimeSaveWorkflow) PlanStart(intent RuntimeSaveIntent, hasEffectiveChanges bool) RuntimeSaveRequestDecision {
	if !hasEffectiveChanges {
		decision := RuntimeSaveRequestDecision{}
		if intent == RuntimeSaveIntentSaveOnly {
			decision.ImmediateStatus = "No changes to save"
		}
		return decision
	}

	return RuntimeSaveRequestDecision{
		StartSave:     true,
		SuccessAction: runtimeSaveSuccessActionForIntent(intent),
	}
}

func (w *RuntimeSaveWorkflow) ResolveResult(successAction RuntimeSaveSuccessAction, affectedRows int, err error) RuntimeSaveResultDecision {
	decision := RuntimeSaveResultDecision{
		ClearPendingSaveAction: true,
	}

	if err != nil {
		decision.StatusMessage = "Error: " + err.Error()
		return decision
	}

	decision.ClearStaging = true
	switch successAction {
	case RuntimeSaveSuccessActionQuitRuntime:
		decision.NextAction = RuntimeSaveResultNextActionQuitRuntime
	case RuntimeSaveSuccessActionRunPendingTransition:
		decision.NextAction = RuntimeSaveResultNextActionRunPendingTransition
	default:
		decision.StatusMessage = fmt.Sprintf("Affected rows: %d", affectedRows)
		decision.NextAction = RuntimeSaveResultNextActionReloadRecords
	}

	return decision
}

func runtimeSaveSuccessActionForIntent(intent RuntimeSaveIntent) RuntimeSaveSuccessAction {
	switch intent {
	case RuntimeSaveIntentSaveAndQuit:
		return RuntimeSaveSuccessActionQuitRuntime
	case RuntimeSaveIntentSaveForPendingTransition:
		return RuntimeSaveSuccessActionRunPendingTransition
	case RuntimeSaveIntentSaveOnly:
		return RuntimeSaveSuccessActionStayInRuntime
	default:
		return RuntimeSaveSuccessActionNone
	}
}
