package registry_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	registrySDK "github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry"
	registrytestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry/testfuncs"
)

func TestAccDataSourceImageTag_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expectedTagName := "test"
	namespaceName := "test-namespace-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isTagDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					# Create a Scaleway registry namespace
					resource "scaleway_registry_namespace" "test" {
						name        = "%s"
						description = "Test namespace for Docker image"
						is_public   = false
					}
				`, namespaceName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["scaleway_registry_namespace.test"]
						if !ok {
							return fmt.Errorf("not found: scaleway_registry_namespace.test")
						}

						endpoint := rs.Primary.Attributes["endpoint"]
						if endpoint == "" {
							return fmt.Errorf("no endpoint found for scaleway_registry_namespace.test")
						}

						return registrytestfuncs.PushImageToRegistry(tt, endpoint, expectedTagName)(s)
					},
				),
			},
			{
				Config: fmt.Sprintf(`
					# Validate the image tag in the Scaleway registry

					data "scaleway_registry_namespace" "test" {
						name = "%s"
					}

					data "scaleway_registry_image" "image" {
  						name = "alpine"
					}

					data "scaleway_registry_image_tag" "tag" {
						name  = "%s"
						image_id = "${data.scaleway_registry_image.image.id}"
					}
				`, namespaceName, expectedTagName),
				Check: resource.ComposeTestCheckFunc(
					isTagPresent(tt, "data.scaleway_registry_image_tag.tag"),
					resource.TestCheckResourceAttr("data.scaleway_registry_image_tag.tag", "name", expectedTagName),
					//resource.TestCheckResourceAttr("scaleway_registry_namespace.test", "name", namespaceName),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_image_tag.tag", "digest"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_image_tag.tag", "created_at"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_image_tag.tag", "updated_at"),
				),
			},
		},
	})
}

func isTagPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := registry.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTag(&registrySDK.GetTagRequest{
			TagID:  id,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

// Note: There is no CheckDestroy for the registry because the image and registry exist in the Scaleway account
// without being managed by Terraform. This makes it challenging to perform a destruction check as Terraform
// has no control over these resources, and their lifecycle is not tracked by Terraform.
func isTagDestroyed(_ *acctest.TestTools) resource.TestCheckFunc {
	return nil
}
