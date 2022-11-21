package scaleway

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
)

func TestAccScalewayFunctionToken_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	expiresAt := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
	if !*UpdateCassettes {
		expiresAt = "2022-10-18T14:04:37+02:00"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionTokenDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_function_namespace main {
						name = "test-function-token-ns"
					}

					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_function_token namespace {
						namespace_id = scaleway_function_namespace.main.id
						expires_at = "%s"
					}

					resource scaleway_function_token function {
						function_id = scaleway_function.main.id
					}
				`, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionTokenExists(tt, "scaleway_function_token.namespace"),
					testAccCheckScalewayFunctionTokenExists(tt, "scaleway_function_token.function"),
					testCheckResourceAttrUUID("scaleway_function_token.namespace", "id"),
					testCheckResourceAttrUUID("scaleway_function_token.function", "id"),
					resource.TestCheckResourceAttrSet("scaleway_function_token.namespace", "token"),
					resource.TestCheckResourceAttrSet("scaleway_function_token.function", "token"),
				),
			},
		},
	})
}

func testAccCheckScalewayFunctionTokenExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetToken(&function.GetTokenRequest{
			TokenID: id,
			Region:  region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFunctionTokenDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_token" {
				continue
			}

			api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteToken(&function.DeleteTokenRequest{
				TokenID: id,
				Region:  region,
			})

			if err == nil {
				return fmt.Errorf("function token (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
