package usecase_test

import (
	"errors"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestRuntimeNavigationWorkflow_PlanTableSwitch_ReturnsDirtyPromptAndPendingAction(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()

	// Act
	plan := workflow.PlanTableSwitch("orders", true, 3)

	// Assert
	if plan.NextAction.Kind != usecase.RuntimeNavigationNextActionNone {
		t.Fatalf("expected no immediate next action, got %v", plan.NextAction.Kind)
	}
	if plan.Pending == nil {
		t.Fatal("expected pending navigation for dirty table switch")
	}
	if plan.Pending.Action.Kind != usecase.RuntimeNavigationActionSwitchTable {
		t.Fatalf("expected pending switch-table action, got %v", plan.Pending.Action.Kind)
	}
	if plan.Pending.Action.TargetTableName != "orders" {
		t.Fatalf("expected pending target table %q, got %q", "orders", plan.Pending.Action.TargetTableName)
	}
	if plan.Prompt == nil {
		t.Fatal("expected dirty table switch prompt")
	}
	if plan.Prompt.Title != "Switch Table" {
		t.Fatalf("expected switch-table prompt title, got %q", plan.Prompt.Title)
	}
	if len(plan.Prompt.Options) != 2 {
		t.Fatalf("expected two prompt options, got %d", len(plan.Prompt.Options))
	}
}

func TestRuntimeNavigationWorkflow_ResolveDecision_ReturnsDiscardSwitchTableAction(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()
	pending := usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind:            usecase.RuntimeNavigationActionSwitchTable,
			TargetTableName: "orders",
		},
	}

	// Act
	result := workflow.ResolveDecision(usecase.DirtyDecisionDiscard, pending)

	// Assert
	if result.Pending != nil {
		t.Fatal("expected discard decision to clear pending navigation")
	}
	if !result.NextAction.ClearDirtyState {
		t.Fatal("expected discard decision to clear dirty state")
	}
	if result.NextAction.Kind != usecase.RuntimeNavigationNextActionSwitchTable {
		t.Fatalf("expected switch-table action, got %v", result.NextAction.Kind)
	}
	if result.NextAction.TargetTableName != "orders" {
		t.Fatalf("expected target table %q, got %q", "orders", result.NextAction.TargetTableName)
	}
}

func TestRuntimeNavigationWorkflow_ResolveDecision_ReturnsStayActionForCancel(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()
	pending := usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind:            usecase.RuntimeNavigationActionSwitchTable,
			TargetTableName: "orders",
		},
	}

	// Act
	result := workflow.ResolveDecision(usecase.DirtyDecisionCancel, pending)

	// Assert
	if result.Pending != nil {
		t.Fatal("expected cancel decision to clear pending navigation")
	}
	if result.NextAction.Kind != usecase.RuntimeNavigationNextActionStayInRuntime {
		t.Fatalf("expected stay-in-runtime action, got %v", result.NextAction.Kind)
	}
	if result.NextAction.ClearDirtyState {
		t.Fatal("expected cancel decision not to clear dirty state")
	}
}

func TestRuntimeNavigationWorkflow_PlanQuit_ReturnsDirtyPromptAndPendingAction(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()

	// Act
	plan := workflow.PlanQuit(true, 2)

	// Assert
	if plan.Pending == nil {
		t.Fatal("expected pending navigation for dirty quit")
	}
	if plan.Pending.Action.Kind != usecase.RuntimeNavigationActionQuitRuntime {
		t.Fatalf("expected pending quit action, got %v", plan.Pending.Action.Kind)
	}
	if plan.Prompt == nil {
		t.Fatal("expected dirty quit prompt")
	}
	if plan.Prompt.Title != "Quit" {
		t.Fatalf("expected quit prompt title, got %q", plan.Prompt.Title)
	}
}

func TestRuntimeNavigationWorkflow_ResolveDecision_ForDatabaseTransitionSupportsSaveDiscardAndCancel(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()
	pending := usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind: usecase.RuntimeNavigationActionOpenDatabase,
			DatabaseTarget: usecase.RuntimeDatabaseTarget{
				Option: usecase.RuntimeDatabaseOption{
					Name:       "analytics",
					ConnString: "/tmp/analytics.sqlite",
					Source:     usecase.RuntimeDatabaseOptionSourceConfig,
				},
				TransitionKind: usecase.RuntimeDatabaseTransitionOpenDifferent,
			},
		},
	}

	// Act
	saveResult := workflow.ResolveDecision(usecase.DirtyDecisionSave, pending)
	discardResult := workflow.ResolveDecision(usecase.DirtyDecisionDiscard, pending)
	cancelResult := workflow.ResolveDecision(usecase.DirtyDecisionCancel, pending)

	// Assert
	if saveResult.NextAction.Kind != usecase.RuntimeNavigationNextActionStartSave {
		t.Fatalf("expected save decision to start save, got %v", saveResult.NextAction.Kind)
	}
	if saveResult.Pending == nil {
		t.Fatal("expected save decision to preserve pending navigation")
	}

	if discardResult.NextAction.Kind != usecase.RuntimeNavigationNextActionOpenDatabase {
		t.Fatalf("expected discard decision to open database, got %v", discardResult.NextAction.Kind)
	}
	if !discardResult.NextAction.ClearDirtyState {
		t.Fatal("expected discard decision to clear dirty state")
	}
	if discardResult.Pending != nil {
		t.Fatal("expected discard decision to clear pending navigation")
	}

	if cancelResult.NextAction.Kind != usecase.RuntimeNavigationNextActionStayInRuntime {
		t.Fatalf("expected cancel decision to stay in runtime, got %v", cancelResult.NextAction.Kind)
	}
	if cancelResult.Pending != nil {
		t.Fatal("expected cancel decision to clear pending navigation")
	}
}

func TestRuntimeNavigationWorkflow_ResolveSaveResult_ExecutesPendingActionOnSuccess(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()
	pending := &usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind: usecase.RuntimeNavigationActionOpenDatabase,
			DatabaseTarget: usecase.RuntimeDatabaseTarget{
				Option: usecase.RuntimeDatabaseOption{
					Name:       "primary",
					ConnString: "/tmp/primary.sqlite",
					Source:     usecase.RuntimeDatabaseOptionSourceConfig,
				},
				TransitionKind: usecase.RuntimeDatabaseTransitionReloadCurrent,
			},
		},
	}

	// Act
	result := workflow.ResolveSaveResult(pending, nil)

	// Assert
	if result.Pending != nil {
		t.Fatal("expected successful save to clear pending navigation")
	}
	if result.NextAction.Kind != usecase.RuntimeNavigationNextActionOpenDatabase {
		t.Fatalf("expected successful save to continue database action, got %v", result.NextAction.Kind)
	}
	if result.NextAction.ClearDirtyState {
		t.Fatal("expected post-save continuation not to re-clear dirty state")
	}
}

func TestRuntimeNavigationWorkflow_ResolveSaveResult_PreservesPendingActionOnSaveFailure(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeNavigationWorkflow()
	pending := &usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind: usecase.RuntimeNavigationActionOpenDatabase,
			DatabaseTarget: usecase.RuntimeDatabaseTarget{
				Option: usecase.RuntimeDatabaseOption{
					Name:       "primary",
					ConnString: "/tmp/primary.sqlite",
					Source:     usecase.RuntimeDatabaseOptionSourceConfig,
				},
				TransitionKind: usecase.RuntimeDatabaseTransitionReloadCurrent,
			},
		},
	}

	// Act
	result := workflow.ResolveSaveResult(pending, errors.New("boom"))

	// Assert
	if result.Pending == nil {
		t.Fatal("expected failed save to preserve pending navigation")
	}
	if result.NextAction.Kind != usecase.RuntimeNavigationNextActionStayInRuntime {
		t.Fatalf("expected failed save to keep runtime active, got %v", result.NextAction.Kind)
	}
}
