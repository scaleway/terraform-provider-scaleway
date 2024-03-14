package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func TestAccScalewayDataSourceSecretVersion_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const (
		secretName            = "dataSourceSecretVersionBasic"
		secretDataDescription = "secret description"
		secretVersionData     = "my_super_secret"
		secretVersionDataV2   = "my_super_secret_v2"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewaySecretVersionDestroy(tt),
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
				`, secretName, secretDataDescription, secretVersionData, secretVersionDataV2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v1", "data", scaleway.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v1", "revision", "1"),

					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v2"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_v2", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_v2", "data", scaleway.Base64Encoded([]byte(secretVersionDataV2))),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceSecretVersion_ByNameSecret(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "dataSourceSecretVersionByNameSecret"
	secretVersionData := "my_super_secret"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewaySecretVersionDestroy(tt),
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
				}
				`, secretName, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.data_by_name", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name", "data", scaleway.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data_by_name", "revision", "1"),
				),
			},
		},
	})
}
