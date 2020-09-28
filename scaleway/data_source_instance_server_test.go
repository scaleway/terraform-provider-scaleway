package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceInstanceServer_Basic(t *testing.T) {
	serverName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "scaleway_instance_server" "main" {
  name 	= "` + serverName + `"
  image = "ubuntu_focal"
  type  = "DEV1-S"
  tags  = [ "terraform-test", "data_scaleway_instance_server", "basic" ]
}

data "scaleway_instance_server" "prod" {
  name = "${scaleway_instance_server.main.name}"
}

data "scaleway_instance_server" "stg" {
  server_id = "${scaleway_instance_server.main.id}"
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("data.scaleway_instance_server.prod"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server.prod", "name", serverName),
					testAccCheckScalewayInstanceServerExists("data.scaleway_instance_server.stg"),
					resource.TestCheckResourceAttr("data.scaleway_instance_server.stg", "name", serverName),
				),
			},
		},
	})
}
