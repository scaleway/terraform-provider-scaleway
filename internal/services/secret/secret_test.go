package secret_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestAccSecret_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretNameBasic"
	updatedName := "secretNameBasicUpdated"
	secretDescription := "secret description"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "%s"
				  tags        = ["devtools", "provider", "terraform"]
				}
				`, secretName, secretDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", secretName),
					resource.TestCheckResourceAttr("scaleway_secret.main", "description", secretDescription),
					resource.TestCheckResourceAttr("scaleway_secret.main", "status", secretSDK.SecretStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.0", "devtools"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.1", "provider"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.2", "terraform"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.#", "3"),
					resource.TestCheckResourceAttrSet("scaleway_secret.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret.main", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "update description"
				  tags        = ["devtools"]
				}
				`, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", updatedName),
					resource.TestCheckResourceAttr("scaleway_secret.main", "description", "update description"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.0", "devtools"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.#", "1"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				}
				`, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", secretName),
					resource.TestCheckResourceAttr("scaleway_secret.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_secret.main", "tags.#", "0"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
		},
	})
}

func TestAccSecret_Path(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_secret" "main" {
				  name = "test-secret-path-secret"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", "test-secret-path-secret"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "path", "/"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
			{
				Config: `
				resource "scaleway_secret" "main" {
				  name = "test-secret-path-secret"
                  path = "/test-secret-path"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", "test-secret-path-secret"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "path", "/test-secret-path"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
			{
				Config: `
				resource "scaleway_secret" "main" {
				  name = "test-secret-path-secret"
                  path = "/test-secret-path/"
				}
				`,
				PlanOnly: true,
			},
			{
				Config: `
				resource "scaleway_secret" "main" {
				  name = "test-secret-path-secret"
                  path = "/test-secret-path-change/"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", "test-secret-path-secret"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "path", "/test-secret-path-change"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
			{
				Config: `
				resource "scaleway_secret" "main" {
				  name = "test-secret-path-secret"
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(tt, "scaleway_secret.main"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "name", "test-secret-path-secret"),
					resource.TestCheckResourceAttr("scaleway_secret.main", "path", "/"),
					acctest.CheckResourceAttrUUID("scaleway_secret.main", "id"),
				),
			},
		},
	})
}

func testAccCheckSecretExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSecret(&secretSDK.GetSecretRequest{
			SecretID: id,
			Region:   region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckSecretDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_secret" {
				continue
			}

			api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetSecret(&secretSDK.GetSecretRequest{
				SecretID: id,
				Region:   region,
			})

			if err == nil {
				return fmt.Errorf("secret (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
