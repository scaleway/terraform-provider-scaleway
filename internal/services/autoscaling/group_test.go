package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	instancetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	lbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb/testfuncs"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccAutoScalingGroupResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	// For now, we need to explicitly define the project_id for the CI's Acceptance Tests to pass.
	// We should find a more appropriate solution in the future.
	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		projectID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isGroupDestroyed(tt),
			instancetestfuncs.IsTemplateDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = %[1]q
						template_id = scaleway_instance_template.main.id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic" ]

						scaling_policy = {
							minimum_size = 3
							maximum_size = 12
							scale_in_cooldown = "5m"
							scale_out_cooldown = "40s"
							scale_in_step = 2
							scale_out_step = 1
							fixed_size = 7
						}
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "name"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.1", "scaleway_autoscaling_group"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.2", "basic"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.minimum_size", "3"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.maximum_size", "12"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_in_cooldown", "5m"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_out_cooldown", "40s"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_in_step", "2"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_out_step", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size", "7"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.cpu_target"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.memory_target"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "status"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_instance_template" "secondary" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic", "second" ]
						server_type = "GP1-L"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = %[1]q
						template_id = scaleway_instance_template.secondary.id
						tags = [ "tf-test", "scaleway_autoscaling_group", "basic", "updated" ]
						name = "tf-test-acc-asg-basic"

						scaling_policy = {
							minimum_size = 1
							maximum_size = 17
							scale_in_cooldown = "7m0s"
							scale_out_cooldown = "30s"
							scale_in_step = 1
							scale_out_step = 4
							fixed_size = 10
						}
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "name", "tf-test-acc-asg-basic"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.secondary", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.#", "4"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.0", "tf-test"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.1", "scaleway_autoscaling_group"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.2", "basic"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.3", "updated"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.minimum_size", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.maximum_size", "17"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_in_cooldown", "7m0s"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_out_cooldown", "30s"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_in_step", "1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.scale_out_step", "4"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size", "10"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.cpu_target"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.memory_target"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "status"),
				),
			},
			{
				ResourceName:      "scaleway_autoscaling_group.main",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"scaling_policy.scale_in_cooldown",
				}, // Import with no reference sets duration fields to their canonical form "1m0s" instead of "1m"
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "main" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_instance_template" "secondary" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic", "second" ]
						server_type = "GP1-L"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = %[1]q
						template_id = scaleway_instance_template.secondary.id
						tags = []
						name = "tf-test-acc-asg-basic-renamed"

						scaling_policy = {
							minimum_size = 1
							maximum_size = 17
							cpu_target = 80
						}
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "name", "tf-test-acc-asg-basic-renamed"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.secondary", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.cpu_target", "80"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "status"),
				),
			},
			{
				ResourceName:      "scaleway_autoscaling_group.main",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"scaling_policy.scale_in_cooldown",
					"scaling_policy.scale_out_cooldown",
				}, // Import with no reference sets duration fields to their canonical form "1m0s" instead of "1m"
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_template" "secondary" {
						project_id = %[1]q
						tags = [ "terraform-test", "scaleway_autoscaling_group", "basic", "second" ]
						server_type = "GP1-L"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = %[1]q
						template_id = scaleway_instance_template.secondary.id
						tags = []
						name = "tf-test-acc-asg-basic-renamed"

						scaling_policy = {
							minimum_size = 1
							maximum_size = 17
							memory_target = 95
						}
					}`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.secondary", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.memory_target", "95"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.cpu_target"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "status"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroupResource_LoadBalancerConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			instancetestfuncs.IsTemplateDestroyed(tt),
			isGroupDestroyed(tt),
			lbchecks.IsLbDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb" "lb" {
						type = "LB-S"
					}

					resource "scaleway_lb_backend" "back0" {
						lb_id            = scaleway_lb.lb.id
						forward_protocol = "http"
						forward_port     = 80
					}

					resource "scaleway_instance_template" "main" {
						project_id = scaleway_lb.lb.project_id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = scaleway_lb.lb.project_id
						template_id = scaleway_instance_template.main.id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						name = "tf-test-acc-asg-lb-config"

						scaling_policy = {
							minimum_size = 0
							maximum_size = 25
							fixed_size = 1
						}

						load_balancer_configuration = {
							load_balancer_id = scaleway_lb.lb.id
							backends = [{
								backend_id = scaleway_lb_backend.back0.id
								address_family = "ipv4"
							}]
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "name", "tf-test-acc-asg-lb-config"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.minimum_size", "0"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.maximum_size", "25"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size", "1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.load_balancer_id", "scaleway_lb.lb", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.backend_id", "scaleway_lb_backend.back0", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.address_family", "ipv4"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.private_network_id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.enabled", "false"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.grace_period"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_autoscaling_group.main", "status"),
				),
			},
			{
				ResourceName:      "scaleway_autoscaling_group.main",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"scaling_policy.scale_in_cooldown",
					"scaling_policy.scale_out_cooldown",
				}, // Import with no reference sets duration fields to their canonical form "1m0s" instead of "1m"
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc" {}
					resource "scaleway_vpc_private_network" "pn" {
						vpc_id = scaleway_vpc.vpc.id
					}

					resource "scaleway_lb" "lb" {
						type = "LB-S"
					}

					resource "scaleway_lb_backend" "back0" {
						lb_id            = scaleway_lb.lb.id
						forward_protocol = "http"
						forward_port     = 80
					}

					resource "scaleway_lb_backend" "back1" {
						lb_id            = scaleway_lb.lb.id
						forward_protocol = "tcp"
						forward_port     = 81
					}

					resource "scaleway_lb_backend" "back2" {
						lb_id            = scaleway_lb.lb.id
						forward_protocol = "http"
						forward_port     = 82
					}

					resource "scaleway_instance_template" "main" {
						project_id = scaleway_lb.lb.project_id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = scaleway_lb.lb.project_id
						template_id = scaleway_instance_template.main.id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						name = "tf-test-acc-asg-lb-config"

						scaling_policy = {
							minimum_size = 0
							maximum_size = 25
							fixed_size = 1
						}

						load_balancer_configuration = {
							load_balancer_id = scaleway_lb.lb.id
							backends = [{
								backend_id = scaleway_lb_backend.back0.id
								address_family = "ipv6"
							},{
								backend_id = scaleway_lb_backend.back1.id
								address_family = "ipv4"
								private_network_id = scaleway_vpc_private_network.pn.id
							},{
								backend_id = scaleway_lb_backend.back2.id
								address_family = "ipv6"
								private_network_id = scaleway_vpc_private_network.pn.id
							}]
							auto_healing = {
								enabled = true
								grace_period = "10m"
							}
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "name", "tf-test-acc-asg-lb-config"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.load_balancer_id", "scaleway_lb.lb", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.#", "3"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.backend_id", "scaleway_lb_backend.back0", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.address_family", "ipv6"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.private_network_id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.1.backend_id", "scaleway_lb_backend.back1", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.1.address_family", "ipv4"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.1.private_network_id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.2.backend_id", "scaleway_lb_backend.back2", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.2.address_family", "ipv6"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.2.private_network_id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.enabled", "true"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.grace_period", "10m"),
				),
			},
			{
				ResourceName:      "scaleway_autoscaling_group.main",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"scaling_policy.scale_in_cooldown",
					"scaling_policy.scale_out_cooldown",
					"load_balancer_configuration.auto_healing.grace_period",
				}, // Import with no reference sets duration fields to their canonical form "1m0s" instead of "1m"
			},
			{
				Config: `
					resource "scaleway_lb" "lb" {
						type = "LB-S"
					}

					resource "scaleway_lb_backend" "back1" {
						lb_id            = scaleway_lb.lb.id
						forward_protocol = "tcp"
						forward_port     = 81
					}

					resource "scaleway_instance_template" "main" {
						project_id = scaleway_lb.lb.project_id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						server_type = "PRO2-M"
					}

					resource "scaleway_autoscaling_group" "main" {
						project_id = scaleway_lb.lb.project_id
						template_id = scaleway_instance_template.main.id
						tags = [ "terraform-test", "scaleway_autoscaling_group", "lb_config" ]
						name = "tf-test-acc-asg-lb-config"

						scaling_policy = {
							minimum_size = 0
							maximum_size = 25
							fixed_size = 1
						}

						load_balancer_configuration = {
							load_balancer_id = scaleway_lb.lb.id
							backends = [{
								backend_id = scaleway_lb_backend.back1.id
								address_family = "ipv4"
							}]
							auto_healing = {
								enabled = false
							}
						}
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					isGroupPresent(tt, "scaleway_autoscaling_group.main"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "name", "tf-test-acc-asg-lb-config"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "template_id", "scaleway_instance_template.main", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.minimum_size", "0"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.maximum_size", "25"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "scaling_policy.fixed_size", "1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.load_balancer_id", "scaleway_lb.lb", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.backend_id", "scaleway_lb_backend.back1", "id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.address_family", "ipv4"),
					resource.TestCheckNoResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.backends.0.private_network_id"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.enabled", "false"),
					resource.TestCheckResourceAttr("scaleway_autoscaling_group.main", "load_balancer_configuration.auto_healing.grace_period", "10m0s"), // TF keeps the last value set
				),
			},
		},
	})
}

func isGroupPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api := autoscaling.NewAPI(tt.Meta.ScwClient())

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetGroup(&autoscaling.GetGroupRequest{
			Zone:    zone,
			GroupID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isGroupDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_autoscaling_group" {
				continue
			}

			api := autoscaling.NewAPI(tt.Meta.ScwClient())

			zone, id, err := zonal.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetGroup(&autoscaling.GetGroupRequest{
				Zone:    zone,
				GroupID: id,
			})
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			return fmt.Errorf("group (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}
