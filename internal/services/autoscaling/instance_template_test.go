package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	autoscalingSDK "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/autoscaling"
)

func TestAccInstanceTemplate_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInstanceTemplateDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "main" {
					  iops       = 5000
					  size_in_gb = 10
					}
					
					resource "scaleway_block_snapshot" "main" {
					  name      = "test-ds-block-snapshot-basic"
					  volume_id = scaleway_block_volume.main.id
					}

					resource "scaleway_autoscaling_instance_template" "main" {
						name = "test-autoscaling-instance-template-basic"
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
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceTemplateExists(tt, "scaleway_autoscaling_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "name", "test-autoscaling-instance-template-basic"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "public_ips_v4_count", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "volumes.0.name", "as-volume"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "volumes.0.volume_type", "sbs"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "volumes.0.boot", "true"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "volumes.0.perf_iops", "5000"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_template.main", "volumes.0.from_snapshot.0.snapshot_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "tags.1", "instance-template"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_instance_template.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_instance_template.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccInstanceTemplate_PrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInstanceTemplateDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
					  name = "TestAccASGInstanceTemplatePrivateNetwork"
					}
					
					resource "scaleway_vpc_private_network" "main" {
					  name   = "TestAccASGInstanceTemplatePrivateNetwork"
					  vpc_id = scaleway_vpc.main.id
					}
					
					resource "scaleway_block_volume" "main" {
					  iops       = 5000
					  size_in_gb = 10
					}
					
					resource "scaleway_block_snapshot" "main" {
					  name      = "TestAccASGInstanceTemplatePrivateNetwork"
					  volume_id = scaleway_block_volume.main.id
					}
					
					resource "scaleway_autoscaling_instance_template" "main" {
					  name            = "TestAccASGInstanceTemplatePrivateNetwork"
					  commercial_type = "PLAY2-MICRO"
					  tags            = ["terraform-test", "basic"]
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceTemplateExists(tt, "scaleway_autoscaling_instance_template.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_instance_template.main", "private_network_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_instance_template.main", "private_network_ids.0", "scaleway_vpc_private_network.main", "id"),
				),
			},
		},
	})
}

func testAccCheckInstanceTemplateExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetInstanceTemplate(&autoscalingSDK.GetInstanceTemplateRequest{
			TemplateID: id,
			Zone:       zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInstanceTemplateDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_autoscaling_instance_template" {
				continue
			}

			api, zone, id, err := autoscaling.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteInstanceTemplate(&autoscalingSDK.DeleteInstanceTemplateRequest{
				TemplateID: id,
				Zone:       zone,
			})
			if err == nil {
				return fmt.Errorf("autoscaling instance template (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
