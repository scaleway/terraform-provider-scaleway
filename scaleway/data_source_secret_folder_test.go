package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceSecretFolder_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewaySecretFolderDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_secret_folder main {
  						name = "test-ds-secret-folder-basic"
					}

					data scaleway_secret_folder find_by_name {
						name = scaleway_secret_folder.main.name
					}

					data scaleway_secret_folder find_by_id {
						folder_id = scaleway_secret_folder.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretFolderExists(tt, "scaleway_secret_folder.main"),

					resource.TestCheckResourceAttrPair("scaleway_secret_folder.main", "name", "data.scaleway_secret_folder.find_by_name", "name"),
					resource.TestCheckResourceAttrPair("scaleway_secret_folder.main", "name", "data.scaleway_secret_folder.find_by_id", "name"),
					resource.TestCheckResourceAttrPair("scaleway_secret_folder.main", "id", "data.scaleway_secret_folder.find_by_name", "id"),
					resource.TestCheckResourceAttrPair("scaleway_secret_folder.main", "id", "data.scaleway_secret_folder.find_by_id", "id"),
				),
			},
		},
	})
}
