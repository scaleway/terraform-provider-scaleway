package documentdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	documentdbSDK "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/documentdb"
)

func TestAccInstance_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "main" {
				  name              = "test-documentdb-instance-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_documentdb_instance", "minimal"]
				  volume_size_in_gb = 20
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					isInstancePresent(tt, "scaleway_documentdb_instance.main"),
					acctest.CheckResourceAttrUUID("scaleway_documentdb_instance.main", "id"),
					resource.TestCheckResourceAttr("scaleway_documentdb_instance.main", "name", "test-documentdb-instance-basic"),
				),
			},
		},
	})
}

func isInstancePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := documentdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetInstance(&documentdbSDK.GetInstanceRequest{
			InstanceID: id,
			Region:     region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isInstanceDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_documentdb_instance" {
				continue
			}

			api, region, id, err := documentdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteInstance(&documentdbSDK.DeleteInstanceRequest{
				InstanceID: id,
				Region:     region,
			})

			if err == nil {
				return fmt.Errorf("documentdb instance (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
