package secret_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceSecret_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "scalewayDataSourceSecret"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%[1]s"
					  description = "DataSourceSecret test description"
					}
					
					data "scaleway_secret" "by_name" {
					  name            = scaleway_secret.main.name
					}
					
					data "scaleway_secret" "by_id" {
					  secret_id       = scaleway_secret.main.id
					}
				`, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "data.scaleway_secret.by_name"),
					resource.TestCheckResourceAttr("data.scaleway_secret.by_name", "name", secretName),

					testAccCheckSecretExists(tt, "data.scaleway_secret.by_id"),
					resource.TestCheckResourceAttr("data.scaleway_secret.by_id", "name", secretName),
				),
			},
		},
	})
}

func TestAccDataSourceSecret_Path(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "main" {
					  name = "test-secret-ds-path"
					  path = "/test-secret-ds-path-path"
					}
					
					data "scaleway_secret" "by_name" {
					  name = scaleway_secret.main.name
					  path = "/test-secret-ds-path-path"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "data.scaleway_secret.by_name"),
					resource.TestCheckResourceAttr("data.scaleway_secret.by_name", "name", "test-secret-ds-path"),
				),
			},
		},
	})
}
