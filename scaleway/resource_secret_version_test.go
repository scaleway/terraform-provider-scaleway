package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
)

func TestAccScalewaySecretVersion_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretVersionNameBasic"
	secretDescription := "secret description"
	secretVersionDescription := "secret version description"
	secretVersionData := "my_super_secret"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewaySecretVersionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "%s"
				  tags        = ["devtools", "provider", "terraform"]
				}
				
				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = "%s"
				}
				`, secretName, secretDescription, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", "version1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secretVersionData),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secret.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "%s"
				  tags        = ["devtools", "provider", "terraform"]
				}
				
				resource "scaleway_secret_version" "v1" {
				  description = "%s"
				  secret_id   = scaleway_secret.main.id
				  data        = "%s"
				}
				`, secretName, secretDescription, secretVersionDescription, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", secretVersionDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secretVersionData),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secret.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "%s"
				  tags        = ["devtools", "provider", "terraform"]
				}
				
				resource "scaleway_secret_version" "v1" {
				  description = "%s"
				  secret_id   = scaleway_secret.main.id
				  data        = "%s"
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data        = "another_secret"
				}
				`, secretName, secretDescription, secretVersionDescription, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", secretVersionDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secretVersionData),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secret.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "created_at"),

					testAccCheckScalewaySecretVersionExists(tt, "scaleway_secret_version.v2"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v2", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "description", "version2"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "data", "another_secret"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "status", secret.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v2", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v2", "created_at"),
				),
			},
		},
	})
}

func testAccCheckScalewaySecretVersionExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, revision, err := secretVersionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSecretVersion(&secret.GetSecretVersionRequest{
			SecretID: id,
			Region:   region,
			Revision: revision,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewaySecretVersionDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_secret_version" {
				continue
			}

			api, region, id, revision, err := secretVersionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetSecretVersion(&secret.GetSecretVersionRequest{
				SecretID: id,
				Region:   region,
				Revision: revision,
			})

			if err == nil {
				return fmt.Errorf("secret version (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
