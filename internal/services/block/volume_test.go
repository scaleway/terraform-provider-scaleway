package block_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccVolume_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-basic"
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "name", "test-block-volume-basic"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "20"),
					resource.TestMatchResourceAttr("scaleway_block_volume.main", "srn", regexp.MustCompile(`^srn://block\..+/zones/.+/volumes/.+$`)),
				),
			},
		},
	})
}

func TestAccVolume_UpdateSize(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var volumeID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-basic"
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "name", "test-block-volume-basic"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "20"),
					acctest.CheckResourceIDPersisted("scaleway_block_volume.main", &volumeID),
				),
			},
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-basic"
						iops = 5000
						size_in_gb = 30
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "name", "test-block-volume-basic"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "30"),
					acctest.CheckResourceIDPersisted("scaleway_block_volume.main", &volumeID),
				),
			},
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-basic"
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "name", "test-block-volume-basic"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "20"),
					acctest.CheckResourceIDChanged("scaleway_block_volume.main", &volumeID),
				),
			},
		},
	})
}

func TestAccVolume_FromSnapshot(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume base {
						name = "test-block-volume-from-snapshot-base"
						iops = 5000
						size_in_gb = 20
					}

					resource scaleway_block_snapshot main {
						name = "test-block-volume-from-snapshot"
						volume_id = scaleway_block_volume.base.id
					}

					resource scaleway_block_volume main {
						name = "test-block-volume-from-snapshot"
						iops = 5000
						snapshot_id = scaleway_block_snapshot.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "snapshot_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "size_in_gb", "scaleway_block_volume.base", "size_in_gb"),
				),
			},
		},
	})
}

func TestAccVolume_FromSnapshotWithSize(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume base {
						name = "test-block-volume-from-snapshot-base"
						iops = 5000
						size_in_gb = 20
					}

					resource scaleway_block_snapshot main {
						name = "test-block-volume-from-snapshot"
						volume_id = scaleway_block_volume.base.id
					}

					resource scaleway_block_volume main {
						name = "test-block-volume-from-snapshot"
						iops = 5000
						snapshot_id = scaleway_block_snapshot.main.id
						size_in_gb = 30
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "snapshot_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "30"),
				),
			},
		},
	})
}

func TestAccVolume_UpdateIops(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-update-iops"
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "iops", "5000"),
				),
			},
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-update-iops"
						iops = 15000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "iops", "15000"),
				),
			},
		},
	})
}

// TestAccVolume_UpdateWithoutName verifies that when a volume is created
// without an explicit name (letting the API auto-generate one), subsequent
// updates do not overwrite the generated name with an empty string.
// This is a regression test for the bug where a computed+optional "name"
// attribute marked as unknown in the plan during an update would cause
// ValueString() to return "" and rename the volume to empty.
func TestAccVolume_UpdateWithoutName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var generatedName string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckResourceAttrSet("scaleway_block_volume.main", "name"),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources["scaleway_block_volume.main"]
						if !ok {
							return fmt.Errorf("resource not found: scaleway_block_volume.main")
						}
						generatedName = rs.Primary.Attributes["name"]
						if generatedName == "" {
							return fmt.Errorf("expected auto-generated name to be non-empty")
						}
						return nil
					},
				),
			},
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 15000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "iops", "15000"),
					func(state *terraform.State) error {
						rs := state.RootModule().Resources["scaleway_block_volume.main"]
						currentName := rs.Primary.Attributes["name"]
						if currentName == "" {
							return fmt.Errorf("name was overwritten to empty string after update (bug regression)")
						}
						if currentName != generatedName {
							return fmt.Errorf("name changed from %q to %q after update without changing name config", generatedName, currentName)
						}
						return nil
					},
				),
			},
		},
	})
}

// TestAccVolume_RemoveSnapshotID verifies that removing snapshot_id from
// config causes the volume to be replaced (recreated), not an inconsistent
// result error. This is a regression test for the bug where the
// LocalityPlanModifier did not trigger RequiresReplace on the value→null
// transition, so Update was called instead, which left snapshot_id in state
// while the plan had it as null — triggering Terraform core's
// "Provider produced inconsistent result after apply" error.
func TestAccVolume_RemoveSnapshotID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var volumeID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume base {
						name = "test-block-volume-remove-snap-base"
						iops = 5000
						size_in_gb = 20
					}

					resource scaleway_block_snapshot snap {
						name = "test-block-volume-remove-snap"
						volume_id = scaleway_block_volume.base.id
					}

					resource scaleway_block_volume main {
						name = "test-block-volume-remove-snap"
						iops = 5000
						snapshot_id = scaleway_block_snapshot.snap.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "snapshot_id", "scaleway_block_snapshot.snap", "id"),
					acctest.CheckResourceIDPersisted("scaleway_block_volume.main", &volumeID),
				),
			},
			{
				Config: `
					resource scaleway_block_volume base {
						name = "test-block-volume-remove-snap-base"
						iops = 5000
						size_in_gb = 20
					}

					resource scaleway_block_snapshot snap {
						name = "test-block-volume-remove-snap"
						volume_id = scaleway_block_volume.base.id
					}

					resource scaleway_block_volume main {
						name = "test-block-volume-remove-snap"
						iops = 15000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					resource.TestCheckNoResourceAttr("scaleway_block_volume.main", "snapshot_id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "iops", "15000"),
					// The volume should be replaced (new id) when snapshot_id is removed.
					acctest.CheckResourceIDChanged("scaleway_block_volume.main", &volumeID),
				),
			},
		},
	})
}
