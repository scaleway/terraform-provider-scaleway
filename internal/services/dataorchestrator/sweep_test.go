package dataorchestrator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_dataorchestrator_workflow", &resource.Sweeper{
		Name: "scaleway_dataorchestrator_workflow",
		F:    testSweepWorkflow,
	})
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepWorkflow(_ string) error {
	return acctest.SweepRegions((&dataorchestratorSDK.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		api := dataorchestratorSDK.NewAPI(scwClient)

		listResp, err := api.ListWorkflowDefinitions(&dataorchestratorSDK.ListWorkflowDefinitionsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list workflow definitions: %w", err)
		}

		for _, def := range listResp.WorkflowDefinitions {
			if !acctest.IsTestResource(def.Name) {
				continue
			}

			err := api.DeleteWorkflowDefinition(&dataorchestratorSDK.DeleteWorkflowDefinitionRequest{
				Region:               region,
				WorkflowDefinitionID: def.ID,
			})
			if err != nil && !httperrors.Is404(err) {
				logging.L.Warningf("sweeper: failed to delete workflow definition %s: %s", def.ID, err)
			}
		}

		return nil
	})
}
