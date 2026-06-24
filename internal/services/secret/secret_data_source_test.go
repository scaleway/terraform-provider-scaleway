package secret_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccDataSourceSecret_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "scalewayDataSourceSecretBasic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			secrettestfuncs.CheckSecretDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceSecret test description"
					}
					
					data "scaleway_secret" "by_name" {
					  name            = scaleway_secret.main.name
					  depends_on	  = [scaleway_secret.main]
					  project_id 	  = scaleway_secret.main.project_id
					}
					
					data "scaleway_secret" "by_id" {
					  secret_id       = scaleway_secret.main.id
					  depends_on	  = [scaleway_secret.main]
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
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		ProtoV6ProviderFactories: tt.ProviderFactories,
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
					  project_id = scaleway_secret.main.project_id
					  depends_on = [scaleway_secret.main]
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

func TestAccDataSourceSecret_WithVersions(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "main" {
						name        = "test-secret-versions"
						description = "Secret with versions"
					}

					resource "scaleway_secret_version" "v1" {
						secret_id = scaleway_secret.main.id
						data_wo   = "first-version-data"
					}

					resource "scaleway_secret_version" "v2" {
						secret_id = scaleway_secret.main.id
						data_wo   = "second-version-data"
					}

					data "scaleway_secret" "with_versions" {
						secret_id  = scaleway_secret.main.id
						depends_on = [scaleway_secret_version.v1, scaleway_secret_version.v2]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "data.scaleway_secret.with_versions"),
					resource.TestCheckResourceAttr("data.scaleway_secret.with_versions", "version_count", "2"),
					resource.TestCheckResourceAttr("data.scaleway_secret.with_versions", "versions.#", "2"),
					resource.TestCheckResourceAttr("data.scaleway_secret.with_versions", "versions.0.revision", "2"),
					resource.TestCheckResourceAttrSet("data.scaleway_secret.with_versions", "versions.0.created_at"),
					resource.TestCheckResourceAttrSet("data.scaleway_secret.with_versions", "versions.0.status"),
					resource.TestCheckResourceAttr("data.scaleway_secret.with_versions", "versions.1.revision", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceSecret_NoVersions(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_secret" "main" {
						name = "test-secret-no-versions"
					}

					data "scaleway_secret" "fresh" {
						secret_id  = scaleway_secret.main.id
						depends_on = [scaleway_secret.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_secret.fresh", "version_count", "0"),
					resource.TestCheckResourceAttr("data.scaleway_secret.fresh", "versions.#", "0"),
				),
			},
		},
	})
}
