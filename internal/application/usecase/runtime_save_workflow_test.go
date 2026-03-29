package usecase_test

import (
	"errors"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestRuntimeSaveWorkflow_PlanRequest_ReturnsNoOpStatusWhenSaveOnlyIsClean(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeSaveWorkflow()

	// Act
	decision := workflow.PlanRequest(usecase.RuntimeSaveIntentSaveOnly, false)

	// Assert
	if decision.StartSave {
		t.Fatal("expected clean save-only request not to start save")
	}
	if decision.ImmediateStatus != "No changes to save" {
		t.Fatalf("expected no-op status, got %q", decision.ImmediateStatus)
	}
	if decision.ImmediateExit {
		t.Fatal("expected clean save-only request not to exit")
	}
	if decision.SuccessAction != usecase.RuntimeSaveSuccessActionNone {
		t.Fatalf("expected no success action, got %v", decision.SuccessAction)
	}
}

func TestRuntimeSaveWorkflow_PlanRequest_ExitsImmediatelyWhenSaveAndQuitIsClean(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeSaveWorkflow()

	// Act
	decision := workflow.PlanRequest(usecase.RuntimeSaveIntentSaveAndQuit, false)

	// Assert
	if decision.StartSave {
		t.Fatal("expected clean save-and-quit request not to start save")
	}
	if !decision.ImmediateExit {
		t.Fatal("expected clean save-and-quit request to exit immediately")
	}
	if decision.ImmediateStatus != "" {
		t.Fatalf("expected no status override, got %q", decision.ImmediateStatus)
	}
	if decision.SuccessAction != usecase.RuntimeSaveSuccessActionNone {
		t.Fatalf("expected no success action, got %v", decision.SuccessAction)
	}
}

func TestRuntimeSaveWorkflow_PlanStart_StartsSaveWithExpectedSuccessActionForEffectiveChanges(t *testing.T) {
	for _, tc := range []struct {
		name          string
		intent        usecase.RuntimeSaveIntent
		successAction usecase.RuntimeSaveSuccessAction
	}{
		{
			name:          "save only",
			intent:        usecase.RuntimeSaveIntentSaveOnly,
			successAction: usecase.RuntimeSaveSuccessActionStayInRuntime,
		},
		{
			name:          "save and quit",
			intent:        usecase.RuntimeSaveIntentSaveAndQuit,
			successAction: usecase.RuntimeSaveSuccessActionQuitRuntime,
		},
		{
			name:          "save for pending transition",
			intent:        usecase.RuntimeSaveIntentSaveForPendingTransition,
			successAction: usecase.RuntimeSaveSuccessActionRunPendingTransition,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			workflow := usecase.NewRuntimeSaveWorkflow()

			// Act
			decision := workflow.PlanStart(tc.intent, true)

			// Assert
			if !decision.StartSave {
				t.Fatal("expected effective changes to start save")
			}
			if decision.ImmediateStatus != "" {
				t.Fatalf("expected no immediate status, got %q", decision.ImmediateStatus)
			}
			if decision.ImmediateExit {
				t.Fatal("expected save start not to exit immediately")
			}
			if decision.SuccessAction != tc.successAction {
				t.Fatalf("expected success action %v, got %v", tc.successAction, decision.SuccessAction)
			}
		})
	}
}

func TestRuntimeSaveWorkflow_PlanStart_SkipsContinuationWhenDirtyStateBuildsNoEffectiveChanges(t *testing.T) {
	for _, tc := range []struct {
		name           string
		intent         usecase.RuntimeSaveIntent
		expectedStatus string
	}{
		{
			name:           "save only",
			intent:         usecase.RuntimeSaveIntentSaveOnly,
			expectedStatus: "No changes to save",
		},
		{
			name:           "save and quit",
			intent:         usecase.RuntimeSaveIntentSaveAndQuit,
			expectedStatus: "",
		},
		{
			name:           "save for pending transition",
			intent:         usecase.RuntimeSaveIntentSaveForPendingTransition,
			expectedStatus: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			workflow := usecase.NewRuntimeSaveWorkflow()

			// Act
			decision := workflow.PlanStart(tc.intent, false)

			// Assert
			if decision.StartSave {
				t.Fatal("expected empty effective changes not to start save")
			}
			if decision.ImmediateExit {
				t.Fatal("expected empty effective changes not to exit immediately")
			}
			if decision.ImmediateStatus != tc.expectedStatus {
				t.Fatalf("expected status %q, got %q", tc.expectedStatus, decision.ImmediateStatus)
			}
			if decision.SuccessAction != usecase.RuntimeSaveSuccessActionNone {
				t.Fatalf("expected no success action, got %v", decision.SuccessAction)
			}
		})
	}
}

func TestRuntimeSaveWorkflow_ResolveResult_ReturnsErrorDecisionWhenSaveFails(t *testing.T) {
	// Arrange
	workflow := usecase.NewRuntimeSaveWorkflow()

	// Act
	decision := workflow.ResolveResult(usecase.RuntimeSaveSuccessActionQuitRuntime, 0, errors.New("boom"))

	// Assert
	if decision.ClearStaging {
		t.Fatal("expected failed save not to clear staging")
	}
	if decision.StatusMessage != "Error: boom" {
		t.Fatalf("expected save error status, got %q", decision.StatusMessage)
	}
	if decision.NextAction != usecase.RuntimeSaveResultNextActionNone {
		t.Fatalf("expected no next action, got %v", decision.NextAction)
	}
	if !decision.ClearPendingSaveAction {
		t.Fatal("expected failed save to clear pending save action")
	}
}

func TestRuntimeSaveWorkflow_ResolveResult_ReturnsAdapterActionForSuccessfulSave(t *testing.T) {
	for _, tc := range []struct {
		name           string
		successAction  usecase.RuntimeSaveSuccessAction
		expectedStatus string
		expectedNext   usecase.RuntimeSaveResultNextAction
	}{
		{
			name:           "stay in runtime",
			successAction:  usecase.RuntimeSaveSuccessActionStayInRuntime,
			expectedStatus: "Affected rows: 3",
			expectedNext:   usecase.RuntimeSaveResultNextActionReloadRecords,
		},
		{
			name:           "quit runtime",
			successAction:  usecase.RuntimeSaveSuccessActionQuitRuntime,
			expectedStatus: "",
			expectedNext:   usecase.RuntimeSaveResultNextActionQuitRuntime,
		},
		{
			name:           "run pending transition",
			successAction:  usecase.RuntimeSaveSuccessActionRunPendingTransition,
			expectedStatus: "",
			expectedNext:   usecase.RuntimeSaveResultNextActionRunPendingTransition,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			workflow := usecase.NewRuntimeSaveWorkflow()

			// Act
			decision := workflow.ResolveResult(tc.successAction, 3, nil)

			// Assert
			if !decision.ClearStaging {
				t.Fatal("expected successful save to clear staging")
			}
			if decision.StatusMessage != tc.expectedStatus {
				t.Fatalf("expected status %q, got %q", tc.expectedStatus, decision.StatusMessage)
			}
			if decision.NextAction != tc.expectedNext {
				t.Fatalf("expected next action %v, got %v", tc.expectedNext, decision.NextAction)
			}
			if !decision.ClearPendingSaveAction {
				t.Fatal("expected successful save to clear pending save action")
			}
		})
	}
}
