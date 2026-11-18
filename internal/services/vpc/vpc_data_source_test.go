package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccDataSourceVPC_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	vpcName := "DataSourceVPC_Basic"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcchecks.CheckVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "vpc01" {
					  name = "%s"
					}`, vpcName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc" "vpc01" {
					  name = "%s"
					}

					data "scaleway_vpc" "by_name" {
						name = "${scaleway_vpc.vpc01.name}"
					}

					data "scaleway_vpc" "by_id" {
						vpc_id = "${scaleway_vpc.vpc01.id}"
					}
				`, vpcName),
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsVPCPresent(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc.by_name", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc.by_id", "name", "scaleway_vpc.vpc01", "name"),
					resource.TestCheckResourceAttrWith("data.scaleway_vpc.by_name", "srn", acctest.CheckSRN("vpc", "vpcs")),
					resource.TestCheckResourceAttrWith("data.scaleway_vpc.by_id", "srn", acctest.CheckSRN("vpc", "vpcs")),
				),
			},
		},
	})
}

func TestAccDataSourceVPC_Default(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_vpc" "default" {
						is_default = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_vpc.default", "id"),
					resource.TestCheckResourceAttr("data.scaleway_vpc.default", "name", "default"),
					resource.TestCheckResourceAttr("data.scaleway_vpc.default", "is_default", "true"),
					resource.TestCheckResourceAttr("data.scaleway_vpc.default", "tags.0", "default"),
					resource.TestCheckResourceAttrWith("data.scaleway_vpc.default", "srn", acctest.CheckSRN("vpc", "vpcs")),
				),
			},
		},
	})
}

func TestAccDataSourceVPC_NoDefaultVPC(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "default" {
					  name = "tf-tests-vpc-no-default-vpc"
					}

					data "scaleway_vpc" "default" {
						project_id = scaleway_account_project.default.id
						is_default = true
					}
				`,
				ExpectError: regexp.MustCompile("no default VPC found in region"),
			},
		},
	})
}
