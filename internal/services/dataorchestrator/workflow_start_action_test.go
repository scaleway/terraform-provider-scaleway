package dataorchestrator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// Pause / resume / terminate actions are registered but not acceptance-tested here:
// they need a long-lived non-terminal run ID, which the minimal setColor DSL never
// produces (runs finish in milliseconds). Revisit when a durable wait step exists.

func isWorkflowRunCreated(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := dataorchestratorSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

		runs, err := api.ListWorkflowRuns(&dataorchestratorSDK.ListWorkflowRunsRequest{
			Region:               region,
			WorkflowDefinitionID: &id,
		}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return fmt.Errorf("failed to list workflow runs: %w", err)
		}

		if len(runs.WorkflowRuns) == 0 {
			return fmt.Errorf("no workflow runs found for definition %s", rs.Primary.ID)
		}

		return nil
	}
}

func TestAccActionWorkflowStart_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionWorkflowStart_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-start"
					  description  = "start"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT

					  lifecycle {
					    action_trigger {
					      events  = [after_create]
					      actions = [action.scaleway_dataorchestrator_workflow_start.main]
					    }
					  }
					}

					action "scaleway_dataorchestrator_workflow_start" "main" {
					  config {
					    workflow_definition_id = scaleway_dataorchestrator_workflow.main.id
					    wait_for_completion    = true
					  }
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					isWorkflowRunCreated(tt, "scaleway_dataorchestrator_workflow.main"),
				),
			},
		},
	})
}
