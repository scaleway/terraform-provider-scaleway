package autoscaling_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInstancePolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInstancePolicyDestroy(tt),
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
					  name            = "TestAccASGInstancePolicy"
					  commercial_type = "PLAY2-MICRO"
					  tags            = ["terraform-test", "instance-template"]
					  volumes {
					    name        = "as-volume"
					    volume_type = "sbs"
					    boot        = true
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
					  name              = "TestAccASGInstancePolicy"
					  action            = "scale_down"
					  type              = "flat_count"
					  value             = 1
					  priority          = 2
					  metric {
					    name               = "cpu scale down"
					    managed_metric     = "managed_metric_instance_cpu"
					    operator           = "operator_less_than"
					    aggregate          = "aggregate_average"
					    sampling_range_min = 5
					    threshold          = 40
					  }
					}
				`,
			},
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
					  name            = "TestAccASGInstancePolicy"
					  commercial_type = "PLAY2-MICRO"
					  tags            = ["terraform-test", "instance-template"]
					  volumes {
					    name        = "as-volume"
					    volume_type = "sbs"
					    boot        = true
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
					  name              = "TestAccASGInstancePolicy"
					  action            = "scale_down"
					  type              = "flat_count"
					  value             = 1
					  priority          = 2
					  metric {
					    name               = "cpu scale down"
					    managed_metric     = "managed_metric_instance_cpu"
					    operator           = "operator_less_than"
					    aggregate          = "aggregate_average"
					    sampling_range_min = 5
					    threshold          = 40
					  }
					}

					data "scaleway_autoscaling_instance_policy" "by_name" {
					  name              = scaleway_autoscaling_instance_policy.main.name
					  instance_group_id = scaleway_autoscaling_instance_group.main.id
					}

					data "scaleway_autoscaling_instance_policy" "by_id" {
					  instance_policy_id = scaleway_autoscaling_instance_policy.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstancePolicyExists(tt, "scaleway_autoscaling_instance_policy.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_policy.by_name", "id", "scaleway_autoscaling_instance_policy.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_policy.by_name", "name", "scaleway_autoscaling_instance_policy.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_policy.by_id", "id", "scaleway_autoscaling_instance_policy.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_policy.by_id", "name", "scaleway_autoscaling_instance_policy.main", "name"),
				),
			},
		},
	})
}
