package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceVPCPublicGateway_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	pgName := "TestAccScalewayDataSourceVPCPublicGateway_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_public_gateway" "main" {
						name = "%s"
						type = "VPC-GW-S"
					}`, pgName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_public_gateway" "main" {
						name = "%s"
						type = "VPC-GW-S"
					}

					data "scaleway_vpc_public_gateway" "pg_test_by_name" {
						name = "${scaleway_vpc_public_gateway.main.name}"
					}

					data "scaleway_vpc_public_gateway" "pg_test_by_id" {
						public_gateway_id = "${scaleway_vpc_public_gateway.main.id}"
					}
				`, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayExists(tt, "scaleway_vpc_public_gateway.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_public_gateway.pg_test_by_name", "name",
						"scaleway_vpc_public_gateway.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_public_gateway.pg_test_by_id", "public_gateway_id",
						"scaleway_vpc_public_gateway.main", "id"),
				),
			},
		},
	})
}
