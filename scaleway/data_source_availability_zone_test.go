package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceAvailabilityZone_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDomainRecordDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data scaleway_availability_zone main {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scaleway_availability_zone.main", "region", "fr-par"),
					resource.TestCheckResourceAttr(
						"data.scaleway_availability_zone.main", "zones.0", "fr-par-1"),
				),
			},
			{
				Config: `
					data scaleway_availability_zone main {
						region = "nl-ams"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.scaleway_availability_zone.main", "region", "nl-ams"),
					resource.TestCheckResourceAttr(
						"data.scaleway_availability_zone.main", "zones.0", "nl-ams-1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_availability_zone.main", "zones.1", "nl-ams-2"),
				),
			},
		},
	})
}
