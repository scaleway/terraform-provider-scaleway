package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccServer_UserData_WithCloudInitAtStart(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_instance_server" "base" {
					name = "tf-acc-server-user-data-with-cloud-init-at-start"
					image = "ubuntu_focal"
					type  = "DEV1-S"
					tags  = [ "terraform-test", "scaleway_instance_server", "user_data_with_cloud_init_at_start" ]

					user_data = {
						foo   = "bar"
						cloud-init =  <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF
					}
				}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.cloud-init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}

func TestAccServer_UserData_WithoutCloudInitAtStart(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// Without cloud-init
				Config: `
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-user-data-without-cloud-init-at-start"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "user_data_without_cloud_init_at_start" ]
						root_volume {
							size_in_gb = 20
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.%", "0"),
				),
			},
			{
				// With cloud-init
				Config: `
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-user-data-without-cloud-init-at-start"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "user_data_without_cloud_init_at_start" ]
						root_volume {
							size_in_gb = 20
						}

						user_data = {
							cloud-init = <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.cloud-init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}
