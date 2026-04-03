package usecase_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestRuntimeNavigationWorkflow_PlanTableSwitch(t *testing.T) {
	t.Parallel()

	workflow := usecase.NewRuntimeNavigationWorkflow()

	tests := []struct {
		name            string
		hasDirty        bool
		changeCount     int
		expectedPending *usecase.PendingRuntimeNavigation
		expectedPrompt  *usecase.RuntimeNavigationDecisionPrompt
		expectedAction  usecase.RuntimeNavigationNextAction
	}{
		{
			name:        "clean state switches immediately",
			hasDirty:    false,
			changeCount: 3,
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind:            usecase.RuntimeNavigationNextActionSwitchTable,
				TargetTableName: "orders",
			},
		},
		{
			name:        "dirty state returns prompt and pending action",
			hasDirty:    true,
			changeCount: 3,
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:            usecase.RuntimeNavigationActionSwitchTable,
					TargetTableName: "orders",
				},
			},
			expectedPrompt: &usecase.RuntimeNavigationDecisionPrompt{
				Title:   "Switch Table",
				Message: "Switching tables will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data?",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and switch table"},
					{ID: usecase.DirtyDecisionCancel, Label: "Continue editing"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plan := workflow.PlanTableSwitch("orders", tc.hasDirty, tc.changeCount)

			assertRuntimeNavigationPlan(t, plan, tc.expectedPending, tc.expectedPrompt, tc.expectedAction)
		})
	}
}

func TestRuntimeNavigationWorkflow_PlanQuit(t *testing.T) {
	t.Parallel()

	workflow := usecase.NewRuntimeNavigationWorkflow()

	tests := []struct {
		name            string
		hasDirty        bool
		changeCount     int
		expectedPending *usecase.PendingRuntimeNavigation
		expectedPrompt  *usecase.RuntimeNavigationDecisionPrompt
		expectedAction  usecase.RuntimeNavigationNextAction
	}{
		{
			name:        "clean state quits immediately",
			hasDirty:    false,
			changeCount: 2,
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind: usecase.RuntimeNavigationNextActionQuitRuntime,
			},
		},
		{
			name:        "dirty state returns prompt and pending action",
			hasDirty:    true,
			changeCount: 2,
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{Kind: usecase.RuntimeNavigationActionQuitRuntime},
			},
			expectedPrompt: &usecase.RuntimeNavigationDecisionPrompt{
				Title:   "Quit",
				Message: "Quitting will cause loss of unsaved data (2 rows). Are you sure you want to discard unsaved data and quit?",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and quit"},
					{ID: usecase.DirtyDecisionCancel, Label: "Continue editing"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plan := workflow.PlanQuit(tc.hasDirty, tc.changeCount)

			assertRuntimeNavigationPlan(t, plan, tc.expectedPending, tc.expectedPrompt, tc.expectedAction)
		})
	}
}

func TestRuntimeNavigationWorkflow_PlanDatabaseTransition(t *testing.T) {
	t.Parallel()

	workflow := usecase.NewRuntimeNavigationWorkflow()
	openTarget := runtimeNavigationTarget("analytics", "/tmp/analytics.sqlite", usecase.RuntimeDatabaseTransitionOpenDifferent)
	reloadTarget := runtimeNavigationTarget("primary", "/tmp/primary.sqlite", usecase.RuntimeDatabaseTransitionReloadCurrent)

	tests := []struct {
		name            string
		target          usecase.RuntimeDatabaseTarget
		hasDirty        bool
		changeCount     int
		expectedPending *usecase.PendingRuntimeNavigation
		expectedPrompt  *usecase.RuntimeNavigationDecisionPrompt
		expectedAction  usecase.RuntimeNavigationNextAction
	}{
		{
			name:        "clean open transition executes immediately",
			target:      openTarget,
			hasDirty:    false,
			changeCount: 3,
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind:           usecase.RuntimeNavigationNextActionOpenDatabase,
				DatabaseTarget: openTarget,
			},
		},
		{
			name:        "dirty open transition prompts with open copy",
			target:      openTarget,
			hasDirty:    true,
			changeCount: 3,
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: openTarget,
				},
			},
			expectedPrompt: &usecase.RuntimeNavigationDecisionPrompt{
				Title:   "Open Database",
				Message: "Opening another database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel.",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionSave, Label: "Save and open database"},
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and open database"},
					{ID: usecase.DirtyDecisionCancel, Label: "Cancel"},
				},
			},
		},
		{
			name:        "dirty reload transition prompts with reload copy",
			target:      reloadTarget,
			hasDirty:    true,
			changeCount: 3,
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: reloadTarget,
				},
			},
			expectedPrompt: &usecase.RuntimeNavigationDecisionPrompt{
				Title:   "Reload Database",
				Message: "Reloading the current database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel.",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionSave, Label: "Save and reload database"},
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and reload database"},
					{ID: usecase.DirtyDecisionCancel, Label: "Cancel"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			plan := workflow.PlanDatabaseTransition(tc.target, tc.hasDirty, tc.changeCount)

			assertRuntimeNavigationPlan(t, plan, tc.expectedPending, tc.expectedPrompt, tc.expectedAction)
		})
	}
}

func TestRuntimeNavigationWorkflow_ResolveDecision(t *testing.T) {
	t.Parallel()

	workflow := usecase.NewRuntimeNavigationWorkflow()
	openTarget := runtimeNavigationTarget("analytics", "/tmp/analytics.sqlite", usecase.RuntimeDatabaseTransitionOpenDifferent)

	tests := []struct {
		name            string
		decisionID      string
		pending         usecase.PendingRuntimeNavigation
		expectedPending *usecase.PendingRuntimeNavigation
		expectedAction  usecase.RuntimeNavigationNextAction
	}{
		{
			name:       "save preserves pending open database action",
			decisionID: usecase.DirtyDecisionSave,
			pending: usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: openTarget,
				},
			},
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: openTarget,
				},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{Kind: usecase.RuntimeNavigationNextActionStartSave},
		},
		{
			name:       "discard executes pending switch table action",
			decisionID: usecase.DirtyDecisionDiscard,
			pending: usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:            usecase.RuntimeNavigationActionSwitchTable,
					TargetTableName: "orders",
				},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind:            usecase.RuntimeNavigationNextActionSwitchTable,
				TargetTableName: "orders",
				ClearDirtyState: true,
			},
		},
		{
			name:       "discard executes pending database transition",
			decisionID: usecase.DirtyDecisionDiscard,
			pending: usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: openTarget,
				},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind:            usecase.RuntimeNavigationNextActionOpenDatabase,
				DatabaseTarget:  openTarget,
				ClearDirtyState: true,
			},
		},
		{
			name:       "cancel keeps runtime active",
			decisionID: usecase.DirtyDecisionCancel,
			pending: usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:            usecase.RuntimeNavigationActionSwitchTable,
					TargetTableName: "orders",
				},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{Kind: usecase.RuntimeNavigationNextActionStayInRuntime},
		},
		{
			name:       "unknown decision keeps runtime active",
			decisionID: "unexpected",
			pending: usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{Kind: usecase.RuntimeNavigationActionQuitRuntime},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{Kind: usecase.RuntimeNavigationNextActionStayInRuntime},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := workflow.ResolveDecision(tc.decisionID, tc.pending)

			assertRuntimeNavigationDecisionResult(t, result, tc.expectedPending, tc.expectedAction)
		})
	}
}

func TestRuntimeNavigationWorkflow_ResolveSaveResult(t *testing.T) {
	t.Parallel()

	workflow := usecase.NewRuntimeNavigationWorkflow()
	openTarget := runtimeNavigationTarget("primary", "/tmp/primary.sqlite", usecase.RuntimeDatabaseTransitionReloadCurrent)
	pending := &usecase.PendingRuntimeNavigation{
		Action: usecase.RuntimeNavigationAction{
			Kind:           usecase.RuntimeNavigationActionOpenDatabase,
			DatabaseTarget: openTarget,
		},
	}

	tests := []struct {
		name            string
		pending         *usecase.PendingRuntimeNavigation
		err             error
		expectedPending *usecase.PendingRuntimeNavigation
		expectedAction  usecase.RuntimeNavigationNextAction
	}{
		{
			name:           "nil pending keeps runtime active",
			pending:        nil,
			expectedAction: usecase.RuntimeNavigationNextAction{Kind: usecase.RuntimeNavigationNextActionStayInRuntime},
		},
		{
			name:    "save failure preserves pending action",
			pending: pending,
			err:     errors.New("boom"),
			expectedPending: &usecase.PendingRuntimeNavigation{
				Action: usecase.RuntimeNavigationAction{
					Kind:           usecase.RuntimeNavigationActionOpenDatabase,
					DatabaseTarget: openTarget,
				},
			},
			expectedAction: usecase.RuntimeNavigationNextAction{Kind: usecase.RuntimeNavigationNextActionStayInRuntime},
		},
		{
			name:    "save success executes pending action",
			pending: pending,
			expectedAction: usecase.RuntimeNavigationNextAction{
				Kind:           usecase.RuntimeNavigationNextActionOpenDatabase,
				DatabaseTarget: openTarget,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := workflow.ResolveSaveResult(tc.pending, tc.err)

			assertRuntimeNavigationDecisionResult(t, result, tc.expectedPending, tc.expectedAction)
		})
	}
}

func runtimeNavigationTarget(name, connString string, transition usecase.RuntimeDatabaseTransitionKind) usecase.RuntimeDatabaseTarget {
	return usecase.RuntimeDatabaseTarget{
		Option: usecase.RuntimeDatabaseOption{
			Name:       name,
			ConnString: connString,
			Source:     usecase.RuntimeDatabaseOptionSourceConfig,
		},
		TransitionKind: transition,
	}
}

func assertRuntimeNavigationPlan(t *testing.T, actual usecase.RuntimeNavigationPlan, expectedPending *usecase.PendingRuntimeNavigation, expectedPrompt *usecase.RuntimeNavigationDecisionPrompt, expectedAction usecase.RuntimeNavigationNextAction) {
	t.Helper()

	assertPendingRuntimeNavigation(t, actual.Pending, expectedPending)
	assertRuntimeNavigationPrompt(t, actual.Prompt, expectedPrompt)

	if actual.NextAction != expectedAction {
		t.Fatalf("expected next action %+v, got %+v", expectedAction, actual.NextAction)
	}
}

func assertRuntimeNavigationDecisionResult(t *testing.T, actual usecase.RuntimeNavigationDecisionResult, expectedPending *usecase.PendingRuntimeNavigation, expectedAction usecase.RuntimeNavigationNextAction) {
	t.Helper()

	assertPendingRuntimeNavigation(t, actual.Pending, expectedPending)

	if actual.NextAction != expectedAction {
		t.Fatalf("expected next action %+v, got %+v", expectedAction, actual.NextAction)
	}
}

func assertPendingRuntimeNavigation(t *testing.T, actual, expected *usecase.PendingRuntimeNavigation) {
	t.Helper()

	switch {
	case expected == nil && actual != nil:
		t.Fatalf("expected no pending navigation, got %+v", *actual)
	case expected != nil && actual == nil:
		t.Fatal("expected pending navigation")
	case expected != nil && actual != nil && !reflect.DeepEqual(*actual, *expected):
		t.Fatalf("expected pending navigation %+v, got %+v", *expected, *actual)
	}
}

func assertRuntimeNavigationPrompt(t *testing.T, actual, expected *usecase.RuntimeNavigationDecisionPrompt) {
	t.Helper()

	switch {
	case expected == nil && actual != nil:
		t.Fatalf("expected no prompt, got %+v", *actual)
	case expected != nil && actual == nil:
		t.Fatal("expected prompt")
	case expected != nil && actual != nil && !reflect.DeepEqual(*actual, *expected):
		t.Fatalf("expected prompt %+v, got %+v", *expected, *actual)
	}
}
