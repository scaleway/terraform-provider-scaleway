package dataorchestrator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceWorkflow_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-ds"
					  description  = "ds"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}

					data "scaleway_dataorchestrator_workflow" "by_id" {
					  workflow_id = scaleway_dataorchestrator_workflow.main.id
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckResourceAttr("data.scaleway_dataorchestrator_workflow.by_id", "name", "tf-tests-dataorchestrator-ds"),
					resource.TestCheckResourceAttr("data.scaleway_dataorchestrator_workflow.by_id", "description", "ds"),
					resource.TestCheckResourceAttrSet("data.scaleway_dataorchestrator_workflow.by_id", "version_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_dataorchestrator_workflow.by_id", "workflow_dsl"),
				),
			},
		},
	})
}
