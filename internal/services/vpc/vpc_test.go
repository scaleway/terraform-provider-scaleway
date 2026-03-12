package vpc_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccVPC_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcchecks.CheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "name", "test-vpc"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "is_default", "false"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "enable_routing", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "enable_custom_routes_propagation", "true"),
				),
			},
			{
				ResourceName:      "scaleway_vpc.vpc01",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPC_WithRegion(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcchecks.CheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
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
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcchecks.CheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
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
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcchecks.CheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "test-vpc-disable-routing"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
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
