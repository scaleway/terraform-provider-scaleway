package jobs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jobsSDK "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
)

func TestAccJobDefinition_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckJobDefinitionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-basic"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-basic"),
				),
			},
		},
	})
}

func TestAccJobDefinition_Timeout(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckJobDefinitionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-timeout"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						timeout = "20m"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-timeout"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "timeout", "20m0s"),
				),
			},
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-timeout"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						timeout = "1h30m"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-timeout"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "timeout", "1h30m0s"),
				),
			},
		},
	})
}

func TestAccJobDefinition_Cron(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckJobDefinitionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-cron"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						cron {
							schedule = "5 4 1 * *"
							timezone = "Europe/Paris"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-cron"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.#", "1"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.0.schedule", "5 4 1 * *"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.0.timezone", "Europe/Paris"),
				),
			},
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-cron"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						cron {
							schedule = "5 5 * * *"
							timezone = "America/Jamaica"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-cron"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.#", "1"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.0.schedule", "5 5 * * *"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.0.timezone", "America/Jamaica"),
				),
			},
			{
				Config: `
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-cron"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-cron"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "cron.#", "0"),
				),
			},
		},
	})
}

func TestAccJobDefinition_SecretReference(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckJobDefinitionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name = "job-secret"
					  path = "/one"
					}
					resource "scaleway_secret_version" "main" {
  					  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}
					locals {
						parts = split("/", scaleway_secret.main.id)
						secret_uuid = local.parts[1]
					}

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "latest"
							file = "/home/dev/env"
						}
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "latest"
							environment = "SOME_ENV"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-secret"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.#", "2"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.file", "/home/dev/env"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.1.environment", "SOME_ENV"),
				),
			},
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name = "job-secret"
					  path = "/one"
					}
					resource "scaleway_secret_version" "main" {
  					  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}
					locals {
						parts = split("/", scaleway_secret.main.id)
						secret_uuid = local.parts[1]
					}

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "latest"
							file = "/home/dev/new_env"
						}
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "latest"
							environment = "SOME_ENV"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-secret"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.#", "2"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.file", "/home/dev/new_env"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.1.environment", "SOME_ENV"),
				),
			},
		},
	})
}

func TestAccJobDefinition_WrongSecretReference(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckJobDefinitionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name    = "job-secret"
					}
					resource "scaleway_secret_version" "main" {
  					  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}
					locals {
						parts = split("/", scaleway_secret.main.id)
						secret_uuid = local.parts[1]
					}

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "1"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`the secret .+ is missing a mount point.+`),
			},
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name    = "job-secret"
					}
					resource "scaleway_secret_version" "main" {
		  			  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}
					locals {
						parts = split("/", scaleway_secret.main.id)
						secret_uuid = local.parts[1]
					}
		
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = local.secret_uuid
							secret_version = "1"
							environment = "SOME_ENV"
							file = "/home/dev/env"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`the secret .+ must have exactly one mount point.+`),
			},
		},
	})
}

func testAccCheckJobDefinitionExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetJobDefinition(&jobsSDK.GetJobDefinitionRequest{
			JobDefinitionID: id,
			Region:          region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckJobDefinitionDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_job_definition" {
				continue
			}

			api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteJobDefinition(&jobsSDK.DeleteJobDefinitionRequest{
				JobDefinitionID: id,
				Region:          region,
			})

			if err == nil {
				return fmt.Errorf("jobs jobdefinition (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
