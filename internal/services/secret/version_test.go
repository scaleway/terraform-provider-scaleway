package secret_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestAccSecretVersion_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretVersionNameBasic"
	secretDescription := "secret description"
	secretVersionDescription := "secret version description"
	secretVersionData := "my_super_secret"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(tt),
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
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", "version1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "revision", "1"),
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
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", secretVersionDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "revision", "1"),
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
				  data        = base64encode("%s")
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data        = "another_secret"
				}
				`, secretName, secretDescription, secretVersionDescription, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", secretVersionDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "revision", "1"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "created_at"),

					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v2"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v2", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "description", "version2"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "data", secret.Base64Encoded([]byte("another_secret"))),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttr("scaleway_secret_version.v2", "revision", "2"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v2", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v2", "created_at"),
				),
			},
		},
	})
}

func TestAccSecretVersion_Type(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretVersionNameType"
	secretVersionData := "{\"key\": \"value\"}"
	secretVersionDataInvalid := "{\"key\": \"value\", \"invalid\": {}}"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  type        = "key_value"
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = %q
				}
				`, secretName, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.v1", "secret_id", "scaleway_secret.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "description", "version1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "data", secret.Base64Encoded([]byte(secretVersionData))),
					resource.TestCheckResourceAttr("scaleway_secret_version.v1", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.v1", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  type        = "key_value"
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data        = %q
				}
				`, secretName, secretVersionDataInvalid),
				ExpectError: regexp.MustCompile("data is wrongly formatted"),
			},
		},
	})
}

func TestAccSecretVersion_DataWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretVersionNameDataWO"
	secretDescription := "secret description"

	secretVersionDataWO := "my_super_secret"
	secretVersionUpdatedDataWO := "my_new_super_secret"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "secretA" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret" "secretB" {
				  name        = "%[1]sB"
				  description = "%[2]sB"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "secretA_version" {
				  description = "%[2]s"
				  secret_id   = scaleway_secret.secretA.id
				  data_wo        = "%[3]s"
				}

				resource "scaleway_secret_version" "secretB_version" {
				  description = "%[2]sB"
				  secret_id   = scaleway_secret.secretB.id
				  data_wo        = "%[3]s"
				}
				`, secretName, secretDescription, secretVersionDataWO),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.secretA_version"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.secretA_version", "secret_id", "scaleway_secret.secretA", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "description", secretDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "revision", "1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretA_version", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretA_version", "created_at"),

					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.secretB_version"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.secretB_version", "secret_id", "scaleway_secret.secretB", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "description", secretDescription+"B"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "revision", "1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretB_version", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretB_version", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "secretA" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret" "secretB" {
				  name        = "%[1]sB"
				  description = "%[2]sB"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "secretA_version" {
				  description = "%[2]s"
				  secret_id   = scaleway_secret.secretA.id
				  data_wo     = "%[3]s"
				}

				resource "scaleway_secret_version" "secretB_version" {
				  description = "%[2]sB"
				  secret_id   = scaleway_secret.secretB.id
				  data_wo     = "%[3]s"
				  revision	  = scaleway_secret.secretB.version_count + 1
				}
				`, secretName, secretDescription, secretVersionUpdatedDataWO),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.secretA_version"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.secretA_version", "secret_id", "scaleway_secret.secretA", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "description", secretDescription),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "revision", "1"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretA_version", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretA_version", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretA_version", "created_at"),

					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.secretB_version"),
					resource.TestCheckResourceAttrPair("scaleway_secret_version.secretB_version", "secret_id", "scaleway_secret.secretB", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "description", secretDescription+"B"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "revision", "2"),
					resource.TestCheckResourceAttr("scaleway_secret_version.secretB_version", "status", secretSDK.SecretVersionStatusEnabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretB_version", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_secret_version.secretB_version", "created_at"),
					// Ensure data_wo (and data) are not in state
					testAccCheckNotInState("scaleway_secret_version.secretB_version", "data_wo"),
					testAccCheckNotInState("scaleway_secret_version.secretB_version", "data"),
				),
			},
		},
	})
}

func TestAccSecretVersion_DataError(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	secretName := "secretVersionNameDataWO"
	secretDescription := "secret description"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%s"
				  description = "%s"
				}

				resource "scaleway_secret_version" "v1" {
				  secret_id   = scaleway_secret.main.id
				}
				`, secretName, secretDescription),
				ExpectError: regexp.MustCompile("one of `data,data_wo` must be specified"),
			},
		},
	})
}

func testAccCheckSecretVersionExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, revision, err := secret.NewVersionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSecretVersion(&secretSDK.GetSecretVersionRequest{
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

func testAccCheckSecretVersionDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_secret_version" {
					continue
				}

				api, region, id, revision, err := secret.NewVersionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				secAPI, _, _, err := secret.NewAPIWithRegionAndID(tt.Meta, fmt.Sprintf("%s/%s", region, id))
				if err == nil {
					sec, err := secAPI.GetSecret(&secretSDK.GetSecretRequest{
						SecretID: id,
						Region:   region,
					})

					switch {
					case err == nil && sec != nil && sec.DeletionRequestedAt != nil:
						// Parent is in scheduled deletion: version will be purged, accept as gone
						continue
					case httperrors.Is404(err):
						continue
					case err != nil:
						return retry.NonRetryableError(err)
					}
				}

				_, err = api.GetSecretVersion(&secretSDK.GetSecretVersionRequest{
					SecretID: id,
					Region:   region,
					Revision: revision,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("secret version (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}

func testAccCheckNotInState(resource, attribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource not found: %s", resource)
		}

		if _, exists := rs.Primary.Attributes[attribute]; exists {
			return fmt.Errorf("%s should not be stored in the state", attribute)
		}

		return nil
	}
}
