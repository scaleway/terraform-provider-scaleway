package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceDocumentDBInstanceLoadBalancer_Basic(t *testing.T) {
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
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-ds-document_db-instance-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}
				
				data "scaleway_document_db_load_balancer_endpoint" "find_by_name" {
				  instance_name = scaleway_document_db_instance.main.name
				}
				
				data "scaleway_document_db_load_balancer_endpoint" "find_by_id" {
				  instance_id = scaleway_document_db_instance.main.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "data.scaleway_document_db_load_balancer_endpoint.find_by_name"),
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "data.scaleway_document_db_load_balancer_endpoint.find_by_id"),
					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "name", "data.scaleway_document_db_load_balancer_endpoint.find_by_name", "instance_name"),
					resource.TestCheckResourceAttrPair("scaleway_document_db_instance.main", "name", "data.scaleway_document_db_load_balancer_endpoint.find_by_id", "instance_name"),
				),
			},
		},
	})
}
