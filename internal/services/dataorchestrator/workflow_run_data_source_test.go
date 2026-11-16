package dataorchestrator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestAccDataSourceWorkflowRun_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	client := meta.ExtractScwClient(tt.Meta)
	api := dataorchestratorSDK.NewAPI(client)
	ctx := tt.T.Context()

	projectID, ok := client.GetDefaultProjectID()
	if !ok {
		t.Fatal("missing default project id")
	}

	dsl := testAccWorkflowMinimalDSL

	def, err := api.CreateWorkflowDefinition(&dataorchestratorSDK.CreateWorkflowDefinitionRequest{
		Region:      scw.RegionFrPar,
		ProjectID:   projectID,
		Name:        "tf-tests-dataorchestrator-run-ds",
		Description: "run ds",
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

	run, err = api.WaitForWorkflowRun(&dataorchestratorSDK.WaitForWorkflowRunRequest{
		Region:        scw.RegionFrPar,
		WorkflowRunID: run.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		t.Fatalf("wait workflow run: %v", err)
	}

	runID := regional.NewIDString(scw.RegionFrPar, run.ID)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dataorchestrator_workflow_run" "by_id" {
					  run_id = %q
					}
				`, runID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_dataorchestrator_workflow_run.by_id", "id", runID),
					resource.TestCheckResourceAttr("data.scaleway_dataorchestrator_workflow_run.by_id", "status", string(dataorchestratorSDK.RunStatusSucceeded)),
					resource.TestCheckResourceAttrSet("data.scaleway_dataorchestrator_workflow_run.by_id", "workflow_definition_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_dataorchestrator_workflow_run.by_id", "workflow_dsl"),
				),
			},
		},
	})
}
