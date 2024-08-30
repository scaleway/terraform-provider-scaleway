package registry_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	registrySDK "github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry"
)

func TestAccDataSourceImageTag_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tagID := "086381ac-24da-476f-9d76-a21ac193b88d"
	imageID := "8f572987-ceb3-4cde-9732-d4febbb821c3"
	expectedTagName := "latest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isTagDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_registry_image_tag" "tag" {
						tag_id  = "` + tagID + `"
						image_id = "` + imageID + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isTagPresent(tt, "data.scaleway_registry_image_tag.tag"),
					resource.TestCheckResourceAttr("data.scaleway_registry_image_tag.tag", "name", expectedTagName),
					resource.TestCheckResourceAttr("data.scaleway_registry_image_tag.tag", "tag_id", tagID),
					resource.TestCheckResourceAttr("data.scaleway_registry_image_tag.tag", "image_id", imageID),
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
