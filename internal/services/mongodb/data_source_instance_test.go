package mongodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccMongoDBInstanceDataSource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_mongodb_instance" "test" {
				  name        = "test-mongodb-instance-both"
				  version     = "7.0"
				  node_type   = "MGDB-PLAY2-NANO"
				  node_number = 1
				  user_name   = "my_initial_user"
				  password    = "thiZ_is_v&ry_s3cret"
				}

				data "scaleway_mongodb_instance" "by_name" {
				  name = scaleway_mongodb_instance.test.name
				}

				data "scaleway_mongodb_instance" "by_id" {
				  instance_id = scaleway_mongodb_instance.test.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_name", "name", "test-mongodb-instance-both"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_name", "version", "7.0"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_name", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_name", "node_number", "1"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_id", "name", "test-mongodb-instance-both"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_id", "version", "7.0"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_id", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.by_id", "node_number", "1"),
				),
			},
		},
	})
}
