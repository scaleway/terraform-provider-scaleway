package rdb_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/stretchr/testify/require"
)

var applyMaintenanceCassetteInstanceID = regexp.MustCompile(`/rdb/v1/regions/([^/]+)/instances/([0-9a-f-]+)`)

func TestAccActionRDBInstanceApplyMaintenance_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBInstanceApplyMaintenance_Basic because action are not yet supported on OpenTofu")
	}

	if *acctest.UpdateCassettes {
		t.Skip("Skipping TestAccActionRDBInstanceApplyMaintenance_Basic: requires manual recording with a pre-provisioned instance")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	instanceRegionalID := testAccRDBApplyMaintenanceInstanceRegionalID(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						node_type      = "db-dev-s"
						tags           = ["terraform-test", "apply-maintenance"]
					}

					import {
						to = scaleway_rdb_instance.main
						id = %q
					}
				`, instanceRegionalID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "id", instanceRegionalID),
					resource.TestCheckResourceAttrSet("scaleway_rdb_instance.main", "maintenances.#"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						node_type      = "db-dev-s"
						tags           = ["terraform-test", "apply-maintenance", "trigger"]

						lifecycle {
							action_trigger {
								events  = [after_update]
								actions = [action.scaleway_rdb_instance_apply_maintenance.main]
							}
						}
					}

					import {
						to = scaleway_rdb_instance.main
						id = %q
					}

					action "scaleway_rdb_instance_apply_maintenance" "main" {
						config {
							instance_id = scaleway_rdb_instance.main.id
							wait        = true
						}
					}
				`, instanceRegionalID),
				Check: resource.ComposeTestCheckFunc(
					isInstanceMaintenanceApplied(tt, "scaleway_rdb_instance.main"),
				),
				Destroy: false,
			},
		},
	})
}

func testAccRDBApplyMaintenanceInstanceRegionalID(t *testing.T) string {
	t.Helper()

	if id := os.Getenv("TF_TEST_RDB_MAINTENANCE_INSTANCE_ID"); id != "" {
		return id
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(wd, "testdata/action-rdb-instance-apply-maintenance-basic.cassette.yaml"))
	require.NoError(t, err)

	matches := applyMaintenanceCassetteInstanceID.FindSubmatch(data)
	require.NotNil(t, matches, "instance id not found in cassette")

	return fmt.Sprintf("%s/%s", matches[1], matches[2])
}

func isInstanceMaintenanceApplied(tt *acctest.TestTools, instanceResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[instanceResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", instanceResourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		instance, err := api.GetInstance(&rdbSDK.GetInstanceRequest{
			Region:     region,
			InstanceID: id,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

		if instance == nil {
			return fmt.Errorf("instance %s not found", id)
		}

		if instance.Status != rdbSDK.InstanceStatusReady {
			return fmt.Errorf("instance %s is not ready after maintenance apply, status: %s", id, instance.Status)
		}

		for _, maintenance := range instance.Maintenances {
			if maintenance.Status == rdbSDK.MaintenanceStatusPending || maintenance.Status == rdbSDK.MaintenanceStatusOngoing {
				return fmt.Errorf("instance %s still has pending or ongoing maintenance", id)
			}
		}

		return nil
	}
}
