package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	autoscalingSDK "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/autoscaling"
)

func TestAccInstanceGroup_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
					  name = "TestAccASGInstanceGroup"
					}
					
					resource "scaleway_vpc_private_network" "main" {
					  name   = "TestAccASGInstanceGroup"
					  vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_block_volume" "main" {
					  iops       = 5000
					  size_in_gb = 10
					}
					
					resource "scaleway_block_snapshot" "main" {
					  name      = "test-ds-block-snapshot-basic"
					  volume_id = scaleway_block_volume.main.id
					}

					resource "scaleway_lb_ip" "main" {}
					resource "scaleway_lb" "main" {
					  ip_id = scaleway_lb_ip.main.id
					  name  = "TestAccASGInstanceGroup"
					  type  = "lb-s"
					  private_network {
						private_network_id = scaleway_vpc_private_network.main.id
					  }
					}
					
					resource "scaleway_lb_backend" "main" {
					  lb_id            = scaleway_lb.main.id
					  forward_protocol = "tcp"
					  forward_port     = 80
					  proxy_protocol   = "none"
					}

					resource "scaleway_autoscaling_instance_template" "main" {
						name = "TestAccASGInstanceGroup"
  						commercial_type  = "PLAY2-MICRO"
					  	tags = ["terraform-test", "instance-template"]
						volumes {
						  name = "as-volume"
						  volume_type = "sbs"
						  boot = true
						  from_snapshot {
						    snapshot_id = scaleway_block_snapshot.main.id
					      }
						  perf_iops = 5000
					    }
  						public_ips_v4_count = 1
					    private_network_ids = [scaleway_vpc_private_network.main.id]
					}

					resource "scaleway_autoscaling_instance_group" "main" {
					  name        = "TestAccASGInstanceGroup"
					  template_id = scaleway_autoscaling_instance_template.main.id
					  tags        = ["terraform-test", "instance-group"]
					  capacity {
						max_replicas   = 5
						min_replicas   = 1
						cooldown_delay = "300"
					  }
					  load_balancer {
						id                 = scaleway_lb.main.id
						backend_ids        = [scaleway_lb_backend.main.id]
						private_network_id = scaleway_vpc_private_network.main.id
					  }
					  delete_servers_on_destroy = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(tt, "scaleway_autoscaling_instance_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "name", "TestAccASGInstanceGroup"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_group.main", "template_id", "scaleway_autoscaling_instance_template.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_group.main", "load_balancer.0.id", "scaleway_lb.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_group.main", "load_balancer.0.backend_ids.0", "scaleway_lb_backend.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_group.main", "load_balancer.0.private_network_id", "scaleway_vpc_private_network.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "tags.1", "instance-group"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "capacity.0.max_replicas", "5"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "capacity.0.min_replicas", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_group.main", "capacity.0.cooldown_delay", "300"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_instance_group.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_instance_group.main", "updated_at"),
				),
			},
		},
	})
}

func testAccCheckInstanceGroupExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetInstanceGroup(&autoscalingSDK.GetInstanceGroupRequest{
			InstanceGroupID: id,
			Zone:            zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInstanceGroupDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_autoscaling_instance_group" {
				continue
			}

			api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteInstanceGroup(&autoscalingSDK.DeleteInstanceGroupRequest{
				InstanceGroupID: id,
				Zone:            zone,
			})
			if err == nil {
				return fmt.Errorf("autoscaling instance group (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
