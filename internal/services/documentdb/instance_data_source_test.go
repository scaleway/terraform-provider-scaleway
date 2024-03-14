package documentdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDocumentDBInstance_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckDocumentDBInstanceDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_documentdb_instance main {
						name = "test-ds-document_db-instance-basic"
						node_type = "docdb-play2-pico"
						engine = "FerretDB-1"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 20
					}

					data scaleway_documentdb_instance find_by_name {
						name = scaleway_documentdb_instance.main.name
					}

					data scaleway_documentdb_instance find_by_id {
						instance_id = scaleway_documentdb_instance.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBInstanceExists(tt, "scaleway_documentdb_instance.main"),

					resource.TestCheckResourceAttrPair("scaleway_documentdb_instance.main", "name", "data.scaleway_documentdb_instance.find_by_name", "name"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_instance.main", "name", "data.scaleway_documentdb_instance.find_by_id", "name"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_instance.main", "id", "data.scaleway_documentdb_instance.find_by_name", "id"),
					resource.TestCheckResourceAttrPair("scaleway_documentdb_instance.main", "id", "data.scaleway_documentdb_instance.find_by_id", "id"),
				),
			},
		},
	})
}
