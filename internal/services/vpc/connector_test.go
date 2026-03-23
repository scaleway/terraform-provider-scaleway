package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
)

func TestAccVPCConnector_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isConnectorDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-connector-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-vpc-connector-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-vpc-connector"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isConnectorPresent(tt, "scaleway_vpc_connector.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "name", "tf-vpc-connector"),
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
					  name = "tf-vpc-connector-source"
					}

					resource "scaleway_vpc" "vpc02" {
					  name = "tf-vpc-connector-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-vpc-connector-updated"
					  vpc_id        = scaleway_vpc.vpc01.id
					  target_vpc_id = scaleway_vpc.vpc02.id
					  tags          = ["terraform", "connector"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isConnectorPresent(tt, "scaleway_vpc_connector.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "name", "tf-vpc-connector-updated"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("scaleway_vpc_connector.main", "tags.1", "connector"),
				),
			},
		},
	})
}

func isConnectorPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetVPCConnector(&vpcSDK.GetVPCConnectorRequest{
			VpcConnectorID: ID,
			Region:         region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isConnectorDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_connector" {
				continue
			}

			vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcAPI.GetVPCConnector(&vpcSDK.GetVPCConnectorRequest{
				VpcConnectorID: ID,
				Region:         region,
			})
			if err == nil {
				return fmt.Errorf("VPC connector (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
