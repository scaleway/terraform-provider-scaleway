package jobs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	jobsSDK "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
)

func isJobRunCreated(tt *acctest.TestTools, jobDefinitionID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, jobDefinitionID)
		if err != nil {
			return err
		}

		jobRuns, err := api.ListJobRuns(&jobsSDK.ListJobRunsRequest{
			Region:          region,
			JobDefinitionID: scw.StringPtr(id),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to list job runs: %w", err)
		}

		if len(jobRuns.JobRuns) == 0 {
			return fmt.Errorf("no job runs found for job definition %s", jobDefinitionID)
		}

		// Check that at least one job run exists and is in a valid state
		for _, jobRun := range jobRuns.JobRuns {
			if jobRun.JobDefinitionID == id {
				return nil
			}
		}

		return fmt.Errorf("no job run found for job definition %s", jobDefinitionID)
	}
}

func TestAccActionJobDefinitionStart_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionJobDefinitionStart_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_job_definition" "main" {
						name         = "test-jobs-action-start"
						cpu_limit    = 120
						memory_limit = 256
						image_uri    = "docker.io/alpine:latest"
						command      = "echo 'Hello World'"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_job_definition_start_action.main]
							}
						}
					}

					action "scaleway_job_definition_start_action" "main" {
						config {
							job_definition_id = scaleway_job_definition.main.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isJobRunCreated(tt, "scaleway_job_definition.main.id"),
				),
			},
		},
	})
}

