package vpc_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccVPCObjectStoragePrivateAccess_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isObjectStoragePrivateAccessDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-vpc-object-storage-private-access"
					}

					resource "scaleway_vpc_private_network" "pn01" {
						name   = "tf-vpc-object-storage-private-access-pn"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_object_storage_private_access" "main" {
						vpc_id              = scaleway_vpc.main.id
						private_network_ids = [scaleway_vpc_private_network.pn01.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isObjectStoragePrivateAccessPresent(tt, "scaleway_vpc_object_storage_private_access.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_vpc_object_storage_private_access.main", "vpc_id",
						"scaleway_vpc.main", "id",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_object_storage_private_access.main", "private_network_ids.#", "1",
					),
					resource.TestCheckResourceAttrSet("scaleway_vpc_object_storage_private_access.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_object_storage_private_access.main", "region"),
				),
			},
			{
				ResourceName:      "scaleway_vpc_object_storage_private_access.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCObjectStoragePrivateAccess_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isObjectStoragePrivateAccessDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-vpc-object-storage-private-access-update"
					}

					resource "scaleway_vpc_private_network" "pn01" {
						name   = "tf-vpc-object-storage-private-access-update-pn"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_object_storage_private_access" "main" {
						vpc_id              = scaleway_vpc.main.id
						private_network_ids = [scaleway_vpc_private_network.pn01.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isObjectStoragePrivateAccessPresent(tt, "scaleway_vpc_object_storage_private_access.main"),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_object_storage_private_access.main", "private_network_ids.#", "1",
					),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-vpc-object-storage-private-access-update"
					}

					resource "scaleway_vpc_private_network" "pn01" {
						name   = "tf-vpc-object-storage-private-access-update-pn"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_private_network" "pn02" {
						name   = "tf-vpc-object-storage-private-access-update-pn-2"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_object_storage_private_access" "main" {
						vpc_id = scaleway_vpc.main.id
						private_network_ids = [
							scaleway_vpc_private_network.pn01.id,
							scaleway_vpc_private_network.pn02.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isObjectStoragePrivateAccessPresent(tt, "scaleway_vpc_object_storage_private_access.main"),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_object_storage_private_access.main", "private_network_ids.#", "2",
					),
				),
			},
		},
	})
}

func TestAccVPCObjectStoragePrivateAccess_RawID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isObjectStoragePrivateAccessDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				// IDs given without the region prefix must be kept as is in the state.
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-vpc-object-storage-private-access-bare-uuids"
					}

					resource "scaleway_vpc_private_network" "pn01" {
						name   = "tf-vpc-object-storage-private-access-bare-uuids-pn"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_object_storage_private_access" "main" {
						vpc_id              = split("/", scaleway_vpc.main.id)[1]
						private_network_ids = [split("/", scaleway_vpc_private_network.pn01.id)[1]]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isObjectStoragePrivateAccessPresent(tt, "scaleway_vpc_object_storage_private_access.main"),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_object_storage_private_access.main", "private_network_ids.#", "1",
					),
					resource.TestCheckResourceAttrWith(
						"scaleway_vpc_object_storage_private_access.main", "vpc_id",
						func(value string) error {
							if strings.Contains(value, "/") {
								return fmt.Errorf("expected a bare UUID, got %q", value)
							}

							return nil
						},
					),
				),
			},
			{
				// switching to regional IDs must not produce a diff
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-vpc-object-storage-private-access-bare-uuids"
					}

					resource "scaleway_vpc_private_network" "pn01" {
						name   = "tf-vpc-object-storage-private-access-bare-uuids-pn"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_vpc_object_storage_private_access" "main" {
						vpc_id              = scaleway_vpc.main.id
						private_network_ids = [scaleway_vpc_private_network.pn01.id]
					}
				`,
				PlanOnly: true,
			},
		},
	})
}

func isObjectStoragePrivateAccessPresent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		region, vpcID, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := vpcSDK.NewAPI(tt.Meta.ScwClient())

		res, err := api.GetVPC(&vpcSDK.GetVPCRequest{
			Region: region,
			VpcID:  vpcID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		if !res.ObjectStoragePrivateAccessEnabled {
			return fmt.Errorf("Object Storage private access is not enabled for VPC %s", vpcID)
		}

		return nil
	}
}

func isObjectStoragePrivateAccessDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := vpcSDK.NewAPI(tt.Meta.ScwClient())

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_object_storage_private_access" {
				continue
			}

			region, vpcID, err := regional.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			res, err := api.GetVPC(&vpcSDK.GetVPCRequest{
				Region: region,
				VpcID:  vpcID,
			}, scw.WithContext(context.Background()))
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			if res.ObjectStoragePrivateAccessEnabled {
				return fmt.Errorf("Object Storage private access %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
