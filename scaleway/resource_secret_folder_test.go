package scaleway

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_secret_folder", &resource.Sweeper{
		Name: "scaleway_secret_folder",
		F:    testSweepSecretFolder,
	})
}

func testSweepSecretFolder(_ string) error {
	return sweepRegions((&secret.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		secretAPI := secret.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the secret folders in (%s)", region)
		listFolders, err := secretAPI.ListFolders(
			&secret.ListFoldersRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing folder in (%s) in sweeper: %s", region, err)
		}

		for _, folder := range listFolders.Folders {
			err := secretAPI.DeleteFolder(&secret.DeleteFolderRequest{
				FolderID: folder.ID,
				Region:   region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting folder in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewaySecretFolder_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewaySecretFolderDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_secret_folder main {
						name = "test-secret-folder-basic"
					}

					resource scaleway_secret_folder subfolder {
						name = "test-secret-folder-basic-subfolder"
						path = scaleway_secret_folder.main.full_path
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecretFolderExists(tt, "scaleway_secret_folder.main"),
					testCheckResourceAttrUUID("scaleway_secret_folder.main", "id"),
					resource.TestCheckResourceAttr("scaleway_secret_folder.main", "name", "test-secret-folder-basic"),
					resource.TestCheckResourceAttr("scaleway_secret_folder.subfolder", "name", "test-secret-folder-basic-subfolder"),
					resource.TestCheckResourceAttr("scaleway_secret_folder.main", "full_path", "/test-secret-folder-basic"),
					resource.TestCheckResourceAttr("scaleway_secret_folder.subfolder", "full_path", "/test-secret-folder-basic/test-secret-folder-basic-subfolder"),
				),
			},
		},
	})
}

func testAccCheckScalewaySecretFolderExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := secretAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = getSecretFolderByID(context.Background(), api, region, id)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewaySecretFolderDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_secret_folder" {
				continue
			}

			api, region, id, err := secretAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteFolder(&secret.DeleteFolderRequest{
				FolderID: id,
				Region:   region,
			})

			if err == nil {
				return fmt.Errorf("secret folder (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
