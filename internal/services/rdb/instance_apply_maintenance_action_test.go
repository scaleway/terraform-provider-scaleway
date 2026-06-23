package rdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
	"github.com/stretchr/testify/require"
)

const testAccRDBApplyMaintenanceRegion = "fr-par"

// testAccRDBApplyMaintenanceInstancePool lists pre-provisioned HashiCorp RDB instances with scheduled maintenance.
// When re-recording cassettes, rotate to the next ID if the current one no longer has applicable maintenance.
var testAccRDBApplyMaintenanceInstancePool = []string{
	"8aaa73d9-e7ea-4f4a-9287-83162b19e8a2",
	"908e1b46-24ef-4182-b2e8-77f379d309b5",
	"1489b5e5-b54f-4f6c-a38e-911e9bfd3cb3",
	"256656ff-4840-4857-b5ed-27a6b722d18c",
	"c6c5e5f1-882f-4691-a2c6-8533ade1ce36",
	"47c063f4-5d92-45b4-9b00-6daa78a6805b",
	"ea607df7-f6af-442c-acbc-b3a4c17b7070",
	"4472df44-f34d-44ca-b428-c26e6a13de8d",
}

func TestAccActionRDBInstanceApplyMaintenance_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBInstanceApplyMaintenance_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	instanceRegionalID := testAccRDBApplyMaintenanceInstanceRegionalID(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						node_type      = "db-dev-s"
						disable_backup = true
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
				PreConfig: func() {
					require.NoError(t, rdbchecks.WaitForApplicableMaintenance(tt, instanceRegionalID))
				},
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "main" {
						node_type      = "db-dev-s"
						disable_backup = true
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

func testAccRDBApplyMaintenanceInstanceRegionalID(tt *acctest.TestTools) string {
	region := scw.Region(testAccRDBApplyMaintenanceRegion)

	if *acctest.UpdateCassettes {
		regionalID, err := rdbchecks.FirstInstanceWithApplicableMaintenance(tt, region, testAccRDBApplyMaintenanceInstancePool)
		require.NoError(tt.T, err)

		return regionalID
	}

	return regional.NewIDString(region, testAccRDBApplyMaintenanceInstancePool[0])
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
