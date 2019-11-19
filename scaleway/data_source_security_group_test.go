package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewaySecurityGroupDataSource_Basic(t *testing.T) {
	dataSourceName := "data.scaleway_security_group.test"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewaySecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayDataSourceSecurityGroupConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecurityGroupExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("test%d", ri)),
					resource.TestCheckResourceAttr(dataSourceName, "description", "public gateway"),
				),
			},
		},
	})
}

func testAccCheckScalewayDataSourceSecurityGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scaleway_security_group" "test" {
  name = "test%d"
  description = "public gateway"
}

data "scaleway_security_group" "test" {
  name = "${scaleway_security_group.test.name}"
}
`, rInt)
}
