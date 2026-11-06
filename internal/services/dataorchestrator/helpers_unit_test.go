package dataorchestrator

import (
	"testing"

	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
)

func TestCurrentWorkflowVersion(t *testing.T) {
	t.Parallel()

	current := &dataorchestratorSDK.WorkflowVersion{ID: "current", IsCurrent: true}
	other := &dataorchestratorSDK.WorkflowVersion{ID: "other", IsCurrent: false}

	got, err := currentWorkflowVersion(&dataorchestratorSDK.WorkflowDefinition{
		ID:       "def",
		Versions: []*dataorchestratorSDK.WorkflowVersion{other, current},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "current" {
		t.Fatalf("got version %q, want current", got.ID)
	}

	also := &dataorchestratorSDK.WorkflowVersion{ID: "also", IsCurrent: false}
	_, err = currentWorkflowVersion(&dataorchestratorSDK.WorkflowDefinition{
		ID:       "def",
		Versions: []*dataorchestratorSDK.WorkflowVersion{other, also},
	})
	if err == nil {
		t.Fatal("expected error for ambiguous versions")
	}
}
