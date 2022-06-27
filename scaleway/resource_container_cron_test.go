package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
)

func TestAccScalewayContainerCron_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerCronDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						name = "my-container-with-cron-tf"
						namespace_id = scaleway_container_namespace.main.id
					}

					resource scaleway_container_cron main {
						container_id = scaleway_container.main.id
						schedule = "5 4 * * *" #cron at 04:05
						args = jsonencode({test = "scw"})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerCronExists(tt, "scaleway_container_cron.main"),
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "schedule", "5 4 * * *"),
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "args", "{\"test\":\"scw\"}"),
				),
			},
		},
	})
}

func TestAccScalewayContainerCron_WithMultiArgs(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerCronDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						name = "my-container-with-cron-tf"
						namespace_id = scaleway_container_namespace.main.id
					}


					resource scaleway_container_cron main {
						container_id = scaleway_container.main.id
						schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
						args = jsonencode(
						{
							address   = {
								city    = "Paris"
								country = "FR"
							}
							age       = 23
							firstName = "John"
							isAlive   = true
							lastName  = "Smith"
						}
                		)
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerCronExists(tt, "scaleway_container_cron.main"),
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "schedule", "5 4 1 * *"),
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "args", "{\"address\":{\"city\":\"Paris\",\"country\":\"FR\"},\"age\":23,\"firstName\":\"John\",\"isAlive\":true,\"lastName\":\"Smith\"}"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
					name = "my-container-with-cron-tf"
						namespace_id = scaleway_container_namespace.main.id
					}

					resource scaleway_container_cron main {
						container_id = scaleway_container.main.id
						schedule = "5 4 * * 1" #cron at 04:05
						args = jsonencode({test = "scw"})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "schedule", "5 4 * * 1"),
					resource.TestCheckResourceAttr("scaleway_container_cron.main", "args", "{\"test\":\"scw\"}"),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerCronExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource container cron not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetCron(&container.GetCronRequest{
			CronID: id,
			Region: region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerCronDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_cron" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteCron(&container.DeleteCronRequest{
				CronID: id,
				Region: region,
			})

			if err == nil {
				return fmt.Errorf("container cron (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
