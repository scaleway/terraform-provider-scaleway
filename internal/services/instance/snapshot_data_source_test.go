package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccDataSourceSnapshot_Basic(t *testing.T) {
	t.Skip("Resources \"scaleway_instance_volume\" and \"scaleway_instance_snapshot\" are depracated")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	snapshotName := "tf-snapshot-ds-basic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancetestfuncs.IsVolumeDestroyed(tt),
			isSnapshotDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_volume" "test" {
						size_in_gb = 2
						type = "b_ssd"
					}

					resource "scaleway_instance_snapshot" "from_volume" {
						name = "%s"
						volume_id = scaleway_instance_volume.test.id
					}`, snapshotName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_volume" "test" {
						size_in_gb = 2
						type = "b_ssd"
					}

					resource "scaleway_instance_snapshot" "from_volume" {
						name = "%s"
						volume_id = scaleway_instance_volume.test.id
					}

					data "scaleway_instance_snapshot" "by_id" {
						snapshot_id = scaleway_instance_snapshot.from_volume.id
					}

					data "scaleway_instance_snapshot" "by_name" {
						name = scaleway_instance_snapshot.from_volume.name
					}`, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					isSnapshotPresent(tt, "data.scaleway_instance_snapshot.by_id"),
					isSnapshotPresent(tt, "data.scaleway_instance_snapshot.by_name"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_snapshot.by_id", "id", "scaleway_instance_snapshot.from_volume", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_snapshot.by_id", "name", "scaleway_instance_snapshot.from_volume", "name"),

					resource.TestCheckResourceAttrPair("data.scaleway_instance_snapshot.by_name", "id", "scaleway_instance_snapshot.from_volume", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_snapshot.by_name", "name", "scaleway_instance_snapshot.from_volume", "name"),
				),
			},
		},
	})
}
