package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceDocumentDBDatabase_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayDocumentDBInstanceDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_document_db_instance main {
						name = "test-ds-document_db-database-basic"
						node_type = "docdb-play2-pico"
						engine = "FerretDB-1"
						is_ha_cluster = false
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 20
					}

					resource scaleway_document_db_database main {
						instance_id = scaleway_document_db_instance.main.id
						name        = "test-ds-document_db-database-basic"
					}

					data scaleway_document_db_database main {
						instance_id = scaleway_document_db_instance.main.id
						name        = scaleway_document_db_database.main.name
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBDatabaseExists(tt, "scaleway_document_db_database.main"),

					resource.TestCheckResourceAttrPair("scaleway_document_db_database.main", "id", "data.scaleway_document_db_database.main", "id"),
				),
			},
		},
	})
}
