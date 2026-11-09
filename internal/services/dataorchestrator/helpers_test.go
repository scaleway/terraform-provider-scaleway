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

func getWorkflowDefinition(tt *acctest.TestTools, resourceName string, state *terraform.State) (*dataorchestratorSDK.WorkflowDefinition, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	region, id, err := regional.ParseID(rs.Primary.ID)
	if err != nil {
		return nil, err
	}

	api := dataorchestratorSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

	return api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: id,
	}, scw.WithContext(tt.T.Context()))
}

func isWorkflowPresent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getWorkflowDefinition(tt, resourceName, state)

		return err
	}
}

func checkWorkflowCronAbsent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		def, err := getWorkflowDefinition(tt, resourceName, state)
		if err != nil {
			return err
		}

		if def.CronSchedule != nil {
			return fmt.Errorf("expected cron_schedule to be cleared on API, got %#v", def.CronSchedule)
		}

		return nil
	}
}

func checkWorkflowCronPresent(tt *acctest.TestTools, resourceName, schedule, timezone string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		def, err := getWorkflowDefinition(tt, resourceName, state)
		if err != nil {
			return err
		}

		if def.CronSchedule == nil {
			return fmt.Errorf("expected cron_schedule on API")
		}

		if def.CronSchedule.Schedule != schedule {
			return fmt.Errorf("cron schedule: got %q want %q", def.CronSchedule.Schedule, schedule)
		}

		if def.CronSchedule.TimezoneIDentifier != timezone {
			return fmt.Errorf("cron timezone: got %q want %q", def.CronSchedule.TimezoneIDentifier, timezone)
		}

		return nil
	}
}

func checkWorkflowEnvAbsent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		def, err := getWorkflowDefinition(tt, resourceName, state)
		if err != nil {
			return err
		}

		if len(def.EnvironmentVariables) != 0 {
			return fmt.Errorf("expected environment_variables cleared on API, got %#v", def.EnvironmentVariables)
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
