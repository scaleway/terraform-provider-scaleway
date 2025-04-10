package jobs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jobsSDK "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = scaleway_secret.main.id
							file = "/home/dev/env"
						}
						secret_reference {
							secret_id = scaleway_secret.main.id
							environment = "SOME_ENV"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-secret"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.secret_version", "latest"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.1.secret_version", "latest"),
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

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = split("/", scaleway_secret.main.id)[1]
							file = "/home/dev/secret_file"
						}
						secret_reference {
							secret_id = scaleway_secret.main.id
							environment = "ANOTHER_ENV"
							secret_version = "1"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-secret"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.#", "2"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.secret_version", "latest"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.1.secret_version", "1"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.file", "/home/dev/secret_file"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.1.environment", "ANOTHER_ENV"),
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

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = scaleway_secret.main.id
							file = "/home/dev/secret_file"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "name", "test-jobs-job-definition-secret"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.#", "1"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.secret_version", "latest"),
					resource.TestCheckResourceAttr("scaleway_job_definition.main", "secret_reference.0.file", "/home/dev/secret_file"),
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
					  name    = "wrong-job-secret"
					}
					resource "scaleway_secret_version" "main" {
  					  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}

					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = scaleway_secret.main.id
						}
					}
				`,
				ExpectError: regexp.MustCompile(`the secret .+ is missing a mount point.+`),
			},
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name    = "wrong-job-secret"
					}
					resource "scaleway_secret_version" "main" {
		  			  secret_id   = scaleway_secret.main.id
					  data        = "your_secret"
					}
		
					resource scaleway_job_definition main {
						name = "test-jobs-job-definition-secret"
						cpu_limit = 120
						memory_limit = 256
						image_uri = "docker.io/alpine:latest"
						secret_reference {
							secret_id = scaleway_secret.main.id
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

func TestCreateJobDefinitionSecret(t *testing.T) {
	jobSecrets := []jobs.JobDefinitionSecret{
		{
			SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
			SecretVersion: "1",
			Environment:   "SOME_ENV",
		},
		{
			SecretID:      regional.NewID("nl-ams", "11111111-1111-1111-1111-111111111111"),
			SecretVersion: "1",
			File:          "/home/dev/env",
		},
	}

	api := jobsSDK.NewAPI(&scw.Client{})
	region := scw.RegionFrPar
	jobID := "22222222-2222-2222-2222-222222222222"

	err := jobs.CreateJobDefinitionSecret(t.Context(), api, jobSecrets, region, jobID)
	assert.ErrorContains(t, err, fmt.Sprintf("the secret id %s is in the region %s, expected %s", jobSecrets[1].SecretID, jobSecrets[1].SecretID.Region, region))
}

func TestDiffJobDefinitionSecrets(t *testing.T) {
	testCases := []struct {
		name             string
		oldSecretRefs    []jobs.JobDefinitionSecret
		newSecretRefs    []jobs.JobDefinitionSecret
		expectedToCreate []jobs.JobDefinitionSecret
		expectedToDelete []jobs.JobDefinitionSecret
	}{
		{
			name: "no changes",
			oldSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
			},
			newSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
			},
			expectedToCreate: []jobs.JobDefinitionSecret{},
			expectedToDelete: []jobs.JobDefinitionSecret{},
		},
		{
			name: "create secret",
			oldSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
			},
			newSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "ANOTHER_ENV",
				},
			},
			expectedToCreate: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "ANOTHER_ENV",
				},
			},
			expectedToDelete: []jobs.JobDefinitionSecret{},
		},
		{
			name: "delete and create secret",
			oldSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "ANOTHER_ENV",
				},
			},
			newSecretRefs: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					File:          "/home/dev/env",
				},
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "ANOTHER_ENV",
				},
			},
			expectedToCreate: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					File:          "/home/dev/env",
				},
			},
			expectedToDelete: []jobs.JobDefinitionSecret{
				{
					SecretID:      regional.NewID("fr-par", "11111111-1111-1111-1111-111111111111"),
					SecretVersion: "1",
					Environment:   "SOME_ENV",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			toCreate, toDelete, err := jobs.DiffJobDefinitionSecrets(testCase.oldSecretRefs, testCase.newSecretRefs)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedToCreate, toCreate)
			assert.Equal(t, testCase.expectedToDelete, toDelete)
		})
	}
}
