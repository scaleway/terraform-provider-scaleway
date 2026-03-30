package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccVPCConnector_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpctestfuncs.CheckConnectorDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-conn-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-conn-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-conn-basic"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpctestfuncs.IsConnectorPresent(tt, "scaleway_vpc_connector.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "name", "tf-conn-basic"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_connector.main", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_connector.main", "target_vpc_id", "scaleway_vpc.vpc02", "id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_connector.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_connector.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_connector.main", "status"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "region", "fr-par"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-conn-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-conn-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-conn-updated"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					  tags          = ["terraform", "connector"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpctestfuncs.IsConnectorPresent(tt, "scaleway_vpc_connector.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "name", "tf-conn-updated"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.1", "connector"),
				),
			},
		},
	})
}
