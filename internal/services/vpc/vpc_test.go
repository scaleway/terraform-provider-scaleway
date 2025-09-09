package vpc_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccVPC_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "name", "test-vpc"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "is_default", "false"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "enable_routing", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "enable_custom_routes_propagation", "true"),
				),
			},
		},
	})
}

func TestAccVPC_WithRegion(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "region", "fr-par"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name   = "test-vpc"
					  region = "nl-ams"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "region", "nl-ams"),
				),
			},
		},
	})
}

func TestAccVPC_WithTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckNoResourceAttr("scaleway_vpc.vpc01", "tags"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					  tags = ["terraform-test", "vpc"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.1", "vpc"),
				),
			},
		},
	})
}

func TestAccVPC_DisableRouting(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc-disable-routing"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "enable_routing", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name           = "test-vpc-disable-routing"
					  enable_routing = false
					}
				`,
				ExpectError: regexp.MustCompile("routing cannot be disabled on this VPC"),
			},
		},
	})
}

func TestAccVPC_ByIdentity(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name   = "test-vpc-import"
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "name", "test-vpc-import"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "region", "fr-par"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(
						"scaleway_vpc.vpc01",
						map[string]knownvalue.Check{
							"id": knownvalue.StringFunc(func(s string) error {
								if !validation.IsUUID(s) {
									return fmt.Errorf("identity.id is not a valid UUID: %s", s)
								}

								return nil
							}),
							"region": knownvalue.StringExact("fr-par"),
						},
					),
				},
			},
			{
				ResourceName:    "scaleway_vpc.vpc01",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func testAccCheckVPCExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetVPC(&vpcSDK.GetVPCRequest{
			VpcID:  ID,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVPCDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, vpctestfuncs.DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc" {
					continue
				}

				vpcAPI, region, id, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = vpcAPI.GetVPC(&vpcSDK.GetVPCRequest{
					Region: region,
					VpcID:  id,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
