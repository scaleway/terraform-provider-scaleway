package sweeper_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

const sweepTestTag = "tf-sweep-test"

// skipIfNoCassette skips the test when the VCR cassette has not been recorded
// yet. Cassettes can only be recorded with real API credentials
// (TF_UPDATE_CASSETTES=true). Without a cassette the VCR framework fails at
// load time, so we guard against that here.
func skipIfNoCassette(t *testing.T, cassetteName string) {
	t.Helper()

	if os.Getenv("TF_UPDATE_CASSETTES") == "true" {
		return
	}

	path := filepath.Join("testdata", cassetteName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("cassette %s not recorded yet; run with TF_UPDATE_CASSETTES=true to record", path)
	}
}

func testAccCheckSecretGone(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_secret" {
				continue
			}

			api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetSecret(&secretSDK.GetSecretRequest{
				SecretID: id,
				Region:   region,
			}, scw.WithContext(ctx))
			if err == nil {
				return fmt.Errorf("secret %s still exists, expected it to be swept", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return fmt.Errorf("unexpected error checking secret %s: %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckSecretExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSecret(&secretSDK.GetSecretRequest{
			SecretID: id,
			Region:   region,
		})

		return err
	}
}

func TestAccActionSweepResources_SecretDryRun(t *testing.T) {
	skipIfNoCassette(t, "action-sweep-resources-secret-dry-run.cassette.yaml")

	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping test because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
						name = "test-sweep-dry-run"
						tags = ["%s"]

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_sweep_resources.dry_run]
							}
						}
					}

					action "scaleway_sweep_resources" "dry_run" {
						config {
							resource_type = "scaleway_secret"
							tags          = ["%s"]
							regions       = ["*"]
							project_ids   = ["*"]
							dry_run       = true
						}
					}
				`, sweepTestTag, sweepTestTag),
				Check: resource.ComposeTestCheckFunc(
					// The secret must still exist: dry-run must not delete anything.
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.0", sweepTestTag),
				),
			},
		},
	})
}

func TestAccActionSweepResources_SecretDelete(t *testing.T) {
	skipIfNoCassette(t, "action-sweep-resources-secret-delete.cassette.yaml")

	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping test because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: create the secret, dry-run sweep after_create.
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
						name = "test-sweep-delete"
						tags = ["%s"]

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_sweep_resources.dry_run]
							}
						}
					}

					action "scaleway_sweep_resources" "dry_run" {
						config {
							resource_type = "scaleway_secret"
							tags          = ["%s"]
							regions       = ["*"]
							project_ids   = ["*"]
							dry_run       = true
						}
					}
				`, sweepTestTag, sweepTestTag),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
				),
			},
			{
				// Step 2: update the secret to trigger after_update, this time
				// with dry_run=false so the sweeper deletes the secret even
				// though it is managed by Terraform state.
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
						name        = "test-sweep-delete"
						description = "trigger sweep delete"
						tags        = ["%s"]

						lifecycle {
							action_trigger {
								events  = [after_update]
								actions = [action.scaleway_sweep_resources.delete]
							}
						}
					}

					action "scaleway_sweep_resources" "delete" {
						config {
							resource_type = "scaleway_secret"
							tags          = ["%s"]
							regions       = ["*"]
							project_ids   = ["*"]
							dry_run       = false
						}
					}
				`, sweepTestTag, sweepTestTag),
				Check: resource.ComposeTestCheckFunc(
					// The secret must be gone: the sweeper deleted it out of band.
					testAccCheckSecretGone(tt),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
