package dataorchestrator_test

import (
	"fmt"
	"regexp"
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
					checkWorkflowCronPresent(tt, "scaleway_dataorchestrator_workflow.main", "0 6 * * *", "Europe/Paris"),
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
					checkWorkflowCronAbsent(tt, "scaleway_dataorchestrator_workflow.main"),
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

func TestAccWorkflow_CreateWithCronThenRemove(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-cron-create"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					  cron_schedule = {
					    schedule            = "0 6 * * *"
					    timezone_identifier = "Europe/Paris"
					    overlap             = "forbid"
					  }
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule.schedule", "0 6 * * *"),
					checkWorkflowCronPresent(tt, "scaleway_dataorchestrator_workflow.main", "0 6 * * *", "Europe/Paris"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-cron-create"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckNoResourceAttr("scaleway_dataorchestrator_workflow.main", "cron_schedule"),
					checkWorkflowCronAbsent(tt, "scaleway_dataorchestrator_workflow.main"),
				),
			},
		},
	})
}

func TestAccWorkflow_EnvVars(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isWorkflowDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-env"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					  environment_variables = {
					    FOO = "bar"
					  }
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckResourceAttr("scaleway_dataorchestrator_workflow.main", "environment_variables.FOO", "bar"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-env"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				Check: resource.ComposeTestCheckFunc(
					isWorkflowPresent(tt, "scaleway_dataorchestrator_workflow.main"),
					resource.TestCheckNoResourceAttr("scaleway_dataorchestrator_workflow.main", "environment_variables.%"),
					checkWorkflowEnvAbsent(tt, "scaleway_dataorchestrator_workflow.main"),
				),
			},
		},
	})
}

func TestAccWorkflow_Validation(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-validation"
					  description  = ""
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				ExpectError: regexp.MustCompile(`(?i)string length must be at least 1`),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-validation"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					  cron_schedule = {
					    schedule            = "not a cron"
					    timezone_identifier = "Europe/Paris"
					  }
					}
				`, testAccWorkflowMinimalDSL),
				ExpectError: regexp.MustCompile(`(?i)Invalid Cron Expression|valid UNIX cron`),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-validation"
					  region       = "fr-par"
					  version_name = "v1-0-0"
					  workflow_dsl = <<-EOT
%s
EOT
					  cron_schedule = {
					    schedule            = "0 6 * * *"
					    timezone_identifier = "Europe/Paris"
					    overlap             = "nope"
					  }
					}
				`, testAccWorkflowMinimalDSL),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dataorchestrator_workflow" "main" {
					  name         = "tf-tests-dataorchestrator-validation"
					  region       = "fr-par"
					  version_name = "bad name!"
					  workflow_dsl = <<-EOT
%s
EOT
					}
				`, testAccWorkflowMinimalDSL),
				ExpectError: regexp.MustCompile(`(?i)letters, digits, hyphens and underscores|Invalid Attribute Value Match`),
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
