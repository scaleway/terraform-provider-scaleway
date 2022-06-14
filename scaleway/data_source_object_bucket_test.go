package scaleway

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceObjectStorage_Basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket")
	// resourceName := "data.scaleway_object_bucket.main"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_object_bucket" "base-01" {
					name = "%s"
					tags = {
						foo = "bar"
					}
				}

				data "scaleway_object_bucket" "selected" {
					name = scaleway_object_bucket.base-01.name
				}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket.selected", "name", bucketName),
				),
			},
		},
	})
}
