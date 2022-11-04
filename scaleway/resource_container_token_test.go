package scaleway

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
)

func TestAccScalewayContainerToken_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	expiresAt := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	if !*UpdateCassettes {
		expiresAt = "2022-10-18T11:35:15+02:00"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerTokenDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_container_namespace main {
						name = "test-container-token-ns"
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
					}

					resource scaleway_container_token namespace {
						namespace_id = scaleway_container_namespace.main.id
						expires_at = "%s"
					}

					resource scaleway_container_token container {
						container_id = scaleway_container.main.id
					}
				`, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerTokenExists(tt, "scaleway_container_token.namespace"),
					testAccCheckScalewayContainerTokenExists(tt, "scaleway_container_token.container"),
					testCheckResourceAttrUUID("scaleway_container_token.namespace", "id"),
					testCheckResourceAttrUUID("scaleway_container_token.container", "id"),
					resource.TestCheckResourceAttrSet("scaleway_container_token.namespace", "token"),
					resource.TestCheckResourceAttrSet("scaleway_container_token.container", "token"),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerTokenExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetToken(&container.GetTokenRequest{
			TokenID: id,
			Region:  region,
		})
		if err != nil {
			return fmt.Errorf("error while getting token: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayContainerTokenDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_token" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteToken(&container.DeleteTokenRequest{
				TokenID: id,
				Region:  region,
			})

			if err == nil {
				return fmt.Errorf("container token (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return fmt.Errorf("error which is not an expected 404: %w", err)
			}
		}

		return nil
	}
}
