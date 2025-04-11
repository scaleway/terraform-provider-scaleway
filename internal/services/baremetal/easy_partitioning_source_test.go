package baremetal_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const (
	offerISData      = "a60ae97c-268c-40cb-af5f-dd276e917ed7"
	osID             = "7d1914e1-f4ab-47fc-bd8c-b3a23143e87a"
	incompatibleOsIS = "4aff4d9d-b1f4-44b0-ab6f-e4711ac11711"
)

func TestAccDataSourceEasyParitioning_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	mountpoint := "/hello"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					`
			data "scaleway_baremetal_easy_partitioning" "test" {
				offer_id = "%s"
				os_id = "%s"
				swap = false
				ext_4_mountpoint = "/hello"
			}`, offerISData, osID),
				Check: resource.ComposeTestCheckFunc(
					// Top-level attributes
					resource.TestCheckResourceAttr("data.scaleway_baremetal_easy_partitioning.test", "ext4_mountpoint", mountpoint),

					// Check at least one disk exists
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "disks.0.device"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "disks.0.partitions.0.label"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "disks.0.partitions.0.number"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "disks.0.partitions.0.size"),

					// Check partition boolean
					resource.TestCheckResourceAttr("data.scaleway_baremetal_easy_partitioning.test", "disks.0.partitions.0.use_all_available_space", "false"),

					// Filesystem checks
					resource.TestCheckResourceAttr("data.scaleway_baremetal_easy_partitioning.test", "filesystems.0.format", "ext4"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_easy_partitioning.test", "filesystems.0.mountpoint", mountpoint),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "filesystems.0.device"),

					// RAID (optional: you can make this conditional if needed)
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "raids.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "raids.0.level"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_easy_partitioning.test", "raids.0.devices.0"),
				),
			},
		},
	})
}
