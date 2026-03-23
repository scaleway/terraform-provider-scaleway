package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceVPCConnector_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isConnectorDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ds-connector-id-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-ds-connector-id-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-ds-connector-id"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					}

					data "scaleway_vpc_connector" "by_id" {
					  connector_id = scaleway_vpc_connector.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isConnectorPresent(tt, "scaleway_vpc_connector.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_connector.by_id", "name",
						"scaleway_vpc_connector.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_connector.by_id", "vpc_id",
						"scaleway_vpc_connector.main", "vpc_id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_connector.by_id", "target_vpc_id",
						"scaleway_vpc_connector.main", "target_vpc_id"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCConnector_ByFilters(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isConnectorDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ds-connector-filter-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-ds-connector-filter-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-ds-connector-filter"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					}

					data "scaleway_vpc_connector" "by_name" {
					  name       = scaleway_vpc_connector.main.name
					  depends_on = [scaleway_vpc_connector.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_connector.by_name", "id",
						"scaleway_vpc_connector.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_connector.by_name", "vpc_id",
						"scaleway_vpc_connector.main", "vpc_id"),
				),
			},
		},
	})
}
