package dataorchestrator_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const testAccWorkflowWaitDSL = `document:
  dsl: 1.0.3
  namespace: tf-tests
  name: tf-tests-workflow-wait
  version: 1.0.0
do:
  - waitAWhile:
      wait:
        minutes: 10
`

func waitWorkflowRunStatus(tt *acctest.TestTools, region scw.Region, runID string, want ...dataorchestratorSDK.RunStatus) (*dataorchestratorSDK.WorkflowRun, error) {
	api := dataorchestratorSDK.NewAPI(meta.ExtractScwClient(tt.Meta))
	wanted := make(map[dataorchestratorSDK.RunStatus]struct{}, len(want))

	for _, status := range want {
		wanted[status] = struct{}{}
	}

	var result *dataorchestratorSDK.WorkflowRun

	err := retry.RetryContext(tt.T.Context(), 10*time.Minute, func() *retry.RetryError {
		run, err := api.GetWorkflowRun(&dataorchestratorSDK.GetWorkflowRunRequest{
			Region:        region,
			WorkflowRunID: runID,
		}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return retry.RetryableError(err)
		}

		if _, ok := wanted[run.Status]; ok {
			result = run

			return nil
		}

		switch run.Status {
		case dataorchestratorSDK.RunStatusSucceeded, dataorchestratorSDK.RunStatusFailed, dataorchestratorSDK.RunStatusInterrupted:
			return retry.NonRetryableError(fmt.Errorf("workflow run %s reached terminal status %q before %v", runID, run.Status, want))
		}

		return retry.RetryableError(fmt.Errorf("waiting for workflow run %s status in %v, got %q", runID, want, run.Status))
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func checkWorkflowRunStatus(tt *acctest.TestTools, runID string, want ...dataorchestratorSDK.RunStatus) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		region, id, err := regional.ParseID(runID)
		if err != nil {
			return err
		}

		_, err = waitWorkflowRunStatus(tt, region, id, want...)

		return err
	}
}

func testAccWorkflowRunControlActionConfig(actionType, runID, triggerInput string) string {
	return fmt.Sprintf(`
					action "scaleway_dataorchestrator_workflow_%s" "main" {
					  config {
					    workflow_run_id = %q
					  }
					}

					resource "terraform_data" "trigger" {
					  input = %q

					  lifecycle {
					    action_trigger {
					      events  = [after_create, after_update]
					      actions = [action.scaleway_dataorchestrator_workflow_%s.main]
					    }
					  }
					}
				`, actionType, runID, triggerInput, actionType)
}

func TestAccActionWorkflowPauseResumeTerminate_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionWorkflowPauseResumeTerminate_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	client := meta.ExtractScwClient(tt.Meta)
	api := dataorchestratorSDK.NewAPI(client)
	ctx := tt.T.Context()

	projectID, ok := client.GetDefaultProjectID()
	if !ok {
		t.Fatal("missing default project id")
	}

	dsl := testAccWorkflowWaitDSL

	def, err := api.CreateWorkflowDefinition(&dataorchestratorSDK.CreateWorkflowDefinitionRequest{
		Region:      scw.RegionFrPar,
		ProjectID:   projectID,
		Name:        "tf-tests-dataorchestrator-run-ctrl",
		Description: "pause resume terminate",
		WorkflowVersion: &dataorchestratorSDK.CreateWorkflowDefinitionRequestWorkflowVersionConfig{
			Name:        "v1-0-0",
			WorkflowDsl: &dsl,
		},
	}, scw.WithContext(ctx))
	if err != nil {
		t.Fatalf("create workflow definition: %v", err)
	}

	t.Cleanup(func() {
		_ = api.DeleteWorkflowDefinition(&dataorchestratorSDK.DeleteWorkflowDefinitionRequest{
			Region:               scw.RegionFrPar,
			WorkflowDefinitionID: def.ID,
		}, scw.WithContext(ctx))
	})

	run, err := api.StartWorkflowDefinition(&dataorchestratorSDK.StartWorkflowDefinitionRequest{
		Region:               scw.RegionFrPar,
		WorkflowDefinitionID: def.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}

	run, err = waitWorkflowRunStatus(tt, scw.RegionFrPar, run.ID, dataorchestratorSDK.RunStatusRunning)
	if err != nil {
		t.Fatalf("wait until running: %v", err)
	}

	runID := regional.NewIDString(scw.RegionFrPar, run.ID)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowRunControlActionConfig("pause", runID, "pause"),
				// Wait tasks often stay in "pausing" rather than reaching "paused".
				Check: checkWorkflowRunStatus(tt, runID, dataorchestratorSDK.RunStatusPaused, dataorchestratorSDK.RunStatusPausing),
			},
			{
				Config: testAccWorkflowRunControlActionConfig("resume", runID, "resume"),
				Check:  checkWorkflowRunStatus(tt, runID, dataorchestratorSDK.RunStatusRunning),
			},
			{
				Config: testAccWorkflowRunControlActionConfig("terminate", runID, "terminate"),
				Check: checkWorkflowRunStatus(
					tt,
					runID,
					dataorchestratorSDK.RunStatusInterrupted,
					dataorchestratorSDK.RunStatusInterrupting,
				),
			},
		},
	})
}
