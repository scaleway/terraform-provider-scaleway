package autoscaling_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInstanceGroup_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInstanceGroupDestroy(tt),
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
					  name            = "TestAccASGInstanceGroup"
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
			},
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
					  name            = "TestAccASGInstanceGroup"
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

					data "scaleway_autoscaling_instance_group" "by_name" {
					  name = scaleway_autoscaling_instance_group.main.name
					}

					data "scaleway_autoscaling_instance_group" "by_id" {
					  instance_group_id = scaleway_autoscaling_instance_group.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(tt, "scaleway_autoscaling_instance_group.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_group.by_name", "id", "scaleway_autoscaling_instance_group.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_group.by_name", "name", "scaleway_autoscaling_instance_group.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_group.by_id", "id", "scaleway_autoscaling_instance_group.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_autoscaling_instance_group.by_id", "name", "scaleway_autoscaling_instance_group.main", "name"),
				),
			},
		},
	})
}
