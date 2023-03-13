package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceSecret_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	secretName := "DataSourceSecret_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewaySecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceSecret test description"
					}

					data "scaleway_secret" "by_name" {
						name = "${scaleway_secret.main.name}"
					}

					data "scaleway_secret" "by_id" {
						secret_id = "${scaleway_secret.main.id}"
					}
				`, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretExists(tt, "data.scaleway_secret.by_name"),
					resource.TestCheckResourceAttr("data.scaleway_secret.by_name", "name", secretName),

					testAccCheckScalewaySecretExists(tt, "data.scaleway_secret.by_id"),
					resource.TestCheckResourceAttr("data.scaleway_secret.by_id", "name", secretName),
				),
			},
		},
	})
}
