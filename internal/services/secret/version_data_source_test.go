package secret_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestAccDataSourceSecretVersion_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const (
		secretName            = "dataSourceSecretVersionBasic"
		secretDataDescription = "secret description"
		secretVersionData     = "my_super_secret"
		secretVersionDataV2   = "my_super_secret_v2"
		secretVersionDataV3   = "my_super_secret_v3"
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[3]s"
				}
				`, secretName, secretDataDescription, secretVersionData),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[3]s"
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[4]s"
				}
				`, secretName, secretDataDescription, secretVersionData, secretVersionDataV2),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[3]s"
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[4]s"
				}

				data "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				}

				data "scaleway_secret_version" "data_v2" {
				  secret_id = scaleway_secret.main.id
				  revision  = "2"
				}

				data "scaleway_secret_version" "data_latest" {
				  secret_id = scaleway_secret.main.id
				  revision  = "latest"
				}
				`, secretName, secretDataDescription, secretVersionData, secretVersionDataV2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v1", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v1", "revision", "1"),

					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v2"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_v2", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v2", "data", secret.Base64Encoded([]byte(secretVersionDataV2))),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_latest", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_latest", "data", secret.Base64Encoded([]byte(secretVersionDataV2))),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[3]s"
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[4]s"
				}

				resource "scaleway_secret_version" "v3" {
				  description = "version3"
				  secret_id   = scaleway_secret.main.id
				  data_wo     = "%[5]s"
				  revision 	  = 3
				  depends_on = [scaleway_secret.main]
				}

				data "scaleway_secret_version" "data_v3" {
				  secret_id = scaleway_secret.main.id
				  revision  = scaleway_secret_version.v3.revision
				  depends_on = [scaleway_secret_version.v3]
				}
				`, secretName, secretDataDescription, secretVersionData, secretVersionDataV2, secretVersionDataV3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v3"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_v3", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v3", "data", secret.Base64Encoded([]byte(secretVersionDataV3))),
				),
			},
		},
	})
}

func TestAccDataSourceSecretVersion_ByNameSecret(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "dsSecretVersionByNameSecret"
	secretVersionData := "my_super_secret"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "main" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[2]s"
				}
				`, secretName, secretVersionData),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name = "%[1]s"
				  tags = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "main" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%[2]s"
				}

				data "scaleway_secret_version" "data_by_name" {
				  secret_name = scaleway_secret.main.name
				  revision    = "1"
				  project_id  = scaleway_secret.main.project_id
				  depends_on = [scaleway_secret_version.main]
				}

				data "scaleway_secret_version" "data_by_name_latest" {
				  secret_name = scaleway_secret.main.name
				  revision    = "latest"
				  project_id  = scaleway_secret.main.project_id
				  depends_on = [scaleway_secret_version.main]
				}
				`, secretName, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_by_name", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name", "revision", "1"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name_latest", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name_latest", "revision", "1")),
			},
		},
	})
}
