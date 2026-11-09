package dataorchestrator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccWorkflow_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-basic"
					  description  = "basic"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					acctest.CheckResourceAttrUUID("scaleway_dataorchestrator_workflow.main", "id"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "name", "tf-tests-dataorchestrator-basic"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "description", "basic"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "version_name", "v1-0-0"),
					resource.TestCheckResourceAttrSet("scaleway_dataorchestrator_workflow.main", "version_id"),
					resource.TestCheckResourceAttrSet("scaleway_dataorchestrator_workflow.main", "workflow_dsl"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-basic"
					  description  = "updated"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					  cron_schedule = {
					    schedule             = "0 6 * * *"
					    timezone_identifier  = "Europe/Paris"
					    overlap              = "forbid"
					  }
					}
				`, testAccWorkflowUpdatedDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "description", "updated"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule.schedule", "0 6 * * *"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule.timezone_identifier", "Europe/Paris"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule.overlap", "forbid"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-basic"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowUpdatedDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckNoResourceAttr("scaleway_dataorchestrator_workflow.main", "description"),
					resource.TestCheckNoResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule"),
				),
			},
			{
				ResourceName:      "scaleway_dataorchestrator_workflow.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkflow_OmitName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckResourceAttrSet("scaleway_dataorchestrator_workflow.main", "name"),
					acctest.CheckResourceAttrUUID("scaleway_dataorchestrator_workflow.main", "id"),
				),
			},
		},
	})
}
