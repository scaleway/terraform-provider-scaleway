package dataorchestrator_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func isWorkflowPresent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := dataorchestratorSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

		_, err = api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
			Region:               region,
			WorkflowDefinitionID: id,
		}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return err
		}

		return nil
	}
}

func isWorkflowDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_dataorchestrator_workflow" {
				continue
			}

			region, id, err := regional.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			api := dataorchestratorSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

			_, err = api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
				Region:               region,
				WorkflowDefinitionID: id,
			}, scw.WithContext(tt.T.Context()))
			if err == nil {
				return fmt.Errorf("workflow definition %s still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

const testAccWorkflowMinimalDSL = `document:
  dsl: 1.0.3
  namespace: tf-tests
  name: tf-tests-workflow
  version: 1.0.0
do:
  - setColor:
      set:
        color: blue
`

const testAccWorkflowUpdatedDSL = `document:
  dsl: 1.0.3
  namespace: tf-tests
  name: tf-tests-workflow
  version: 1.0.1
do:
  - setColor:
      set:
        color: green
`
