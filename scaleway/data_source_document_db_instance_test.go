package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceDocumentDBInstance_Basic(t *testing.T) {
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
						name = "test-ds-document_db-instance-basic"
						node_type = "docdb-play2-pico"
						engine = "FerretDB-1.0.0"
						is_ha_cluster = false
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_document_db_instance", "minimal" ]
						volume_size_in_gb = 20
					}

					data scaleway_document_db_instance find_by_name {
						name = scaleway_document_db_instance.main.name
					}

					data scaleway_document_db_instance find_by_id {
						instance_id = scaleway_document_db_instance.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceExists(tt, "scaleway_document_db_instance.main"),

					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "name", "data.scaleway_document_db_instance.find_by_name", "name"),
					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "name", "data.scaleway_document_db_instance.find_by_id", "name"),
					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "id", "data.scaleway_document_db_instance.find_by_name", "id"),
					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "id", "data.scaleway_document_db_instance.find_by_id", "id"),
				),
			},
		},
	})
}
