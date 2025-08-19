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

func TestAccInstancePolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckInstancePolicyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
					  name = "TestAccASGInstancePolicy"
					}
					
					resource "scaleway_vpc_private_network" "main" {
					  name   = "TestAccASGInstancePolicy"
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
					  name  = "TestAccASGInstanceTemplatePrivateNetwork"
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
						name = "TestAccASGInstancePolicy"
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
					  name        = "TestAccASGInstancePolicy"
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

					resource "scaleway_autoscaling_instance_policy" "main" {
						instance_group_id = scaleway_autoscaling_instance_group.main.id
						name = "TestAccASGInstancePolicy"
						action = "scale_down"
						type = "flat_count"
						value = 1
						priority = 2
						metric {
						  name = "cpu scale down"
						  managed_metric = "managed_metric_instance_cpu"
						  operator = "operator_less_than"
						  aggregate = "aggregate_average"
						  sampling_range_min = 5
                          threshold = 40
                        }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstancePolicyExists(tt, "scaleway_autoscaling_instance_policy.main"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_policy.main", "instance_group_id", "scaleway_autoscaling_instance_group.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "name", "TestAccASGInstancePolicy"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "action", "scale_down"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "type", "flat_count"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "value", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "priority", "2"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.name", "cpu scale down"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.managed_metric", "managed_metric_instance_cpu"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.operator", "operator_less_than"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.aggregate", "aggregate_average"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.sampling_range_min", "5"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_policy.main", "metric.0.threshold", "40"),
				),
			},
		},
	})
}

func testAccCheckInstancePolicyExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetInstancePolicy(&autoscalingSDK.GetInstancePolicyRequest{
			PolicyID: id,
			Zone:     zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInstancePolicyDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_autoscaling_instance_policy" {
				continue
			}

			api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteInstancePolicy(&autoscalingSDK.DeleteInstancePolicyRequest{
				PolicyID: id,
				Zone:     zone,
			})
			if err == nil {
				return fmt.Errorf("autoscaling instance policy (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
