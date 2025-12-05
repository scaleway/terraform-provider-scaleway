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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
)

func isJobRunCreated(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
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
			return fmt.Errorf("no job runs found for job definition %s", rs.Primary.ID)
		}

		// Check that at least one job run exists and is in a valid state
		for _, jobRun := range jobRuns.JobRuns {
			if jobRun.JobDefinitionID == id {
				return nil
			}
		}

		return fmt.Errorf("no job run found for job definition %s", rs.Primary.ID)
	}
}

func testAccCheckJobDefinitionDestroyIgnoringRunningJobs(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_job_definition" {
				continue
			}

			api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			// List all job runs for this job definition
			jobRuns, err := api.ListJobRuns(&jobsSDK.ListJobRunsRequest{
				Region:          region,
				JobDefinitionID: scw.StringPtr(id),
			}, scw.WithContext(context.Background()))
			if err != nil {
				return fmt.Errorf("failed to list job runs: %w", err)
			}

			// Stop all running or queued job runs and collect their IDs
			var jobRunIDsToWait []string

			for _, jobRun := range jobRuns.JobRuns {
				if jobRun.State == jobsSDK.JobRunStateQueued || jobRun.State == jobsSDK.JobRunStateRunning {
					_, err := api.StopJobRun(&jobsSDK.StopJobRunRequest{
						JobRunID: jobRun.ID,
						Region:   region,
					}, scw.WithContext(context.Background()))
					if err != nil && !httperrors.Is404(err) {
						return fmt.Errorf("failed to stop job run %s: %w", jobRun.ID, err)
					}

					jobRunIDsToWait = append(jobRunIDsToWait, jobRun.ID)
				}
			}

			// Wait for all stopped job runs to terminate
			for _, jobRunID := range jobRunIDsToWait {
				_, err := api.WaitForJobRun(&jobsSDK.WaitForJobRunRequest{
					JobRunID: jobRunID,
					Region:   region,
				}, scw.WithContext(context.Background()))
				if err != nil && !httperrors.Is404(err) {
					return fmt.Errorf("failed to wait for job run %s: %w", jobRunID, err)
				}
			}

			// Now delete the job definition (Terraform may have failed to delete it due to running job runs)
			err = api.DeleteJobDefinition(&jobsSDK.DeleteJobDefinitionRequest{
				JobDefinitionID: id,
				Region:          region,
			})
			if err == nil {
				// Successfully deleted, resource is destroyed
				continue
			}

			// If 404, the resource doesn't exist (already deleted or never existed)
			if httperrors.Is404(err) {
				continue
			}

			// Other errors should be returned
			return fmt.Errorf("failed to delete job definition %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func cleanupJobRuns(tt *acctest.TestTools, jobDefinitionID string) {
	api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, jobDefinitionID)
	if err != nil {
		return
	}

	jobRuns, err := api.ListJobRuns(&jobsSDK.ListJobRunsRequest{
		Region:          region,
		JobDefinitionID: scw.StringPtr(id),
	}, scw.WithContext(context.Background()))
	if err != nil {
		return
	}

	for _, jobRun := range jobRuns.JobRuns {
		if jobRun.State == jobsSDK.JobRunStateQueued || jobRun.State == jobsSDK.JobRunStateRunning {
			_, _ = api.StopJobRun(&jobsSDK.StopJobRunRequest{
				JobRunID: jobRun.ID,
				Region:   region,
			}, scw.WithContext(context.Background()))
		}
	}

	for _, jobRun := range jobRuns.JobRuns {
		if jobRun.State == jobsSDK.JobRunStateQueued || jobRun.State == jobsSDK.JobRunStateRunning {
			_, _ = api.WaitForJobRun(&jobsSDK.WaitForJobRunRequest{
				JobRunID: jobRun.ID,
				Region:   region,
			}, scw.WithContext(context.Background()))
		}
	}

	// Try to delete the job definition after cleaning up job runs
	_ = api.DeleteJobDefinition(&jobsSDK.DeleteJobDefinitionRequest{
		JobDefinitionID: id,
		Region:          region,
	})
}

func TestAccActionJobDefinitionStart_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionJobDefinitionStart_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var jobDefinitionID string

	defer func() {
		if jobDefinitionID != "" {
			cleanupJobRuns(tt, jobDefinitionID)
		}
	}()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroyIgnoringRunningJobs(tt),
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
					isJobRunCreated(tt, "scaleway_job_definition.main"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["scaleway_job_definition.main"]
						if ok {
							jobDefinitionID = rs.Primary.ID
						}

						return nil
					},
				),
			},
		},
	})
}
