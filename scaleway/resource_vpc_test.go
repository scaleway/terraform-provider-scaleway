package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_vpc", &resource.Sweeper{
		Name:         "scaleway_vpc",
		F:            testSweepVPC,
		Dependencies: []string{"scaleway_vpc_private_network"},
	})
}

func testSweepVPC(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		vpcAPI := vpc.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting the VPCs in (%s)", region)

		listVPCs, err := vpcAPI.ListVPCs(&vpc.ListVPCsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing secrets in (%s) in sweeper: %s", region, err)
		}

		for _, v := range listVPCs.Vpcs {
			if v.IsDefault {
				continue
			}
			err := vpcAPI.DeleteVPC(&vpc.DeleteVPCRequest{
				VpcID:  v.ID,
				Region: region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting VPC in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayVPC_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "name", "test-vpc"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "is_default", "false"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc.vpc01", "updated_at"),
				),
			},
		},
	})
}

func TestAccScalewayVPC_WithRegion(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "region", "fr-par"),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
						region = "nl-ams"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "region", "nl-ams"),
				),
			},
		},
	})
}

func TestAccScalewayVPC_WithTags(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckNoResourceAttr("scaleway_vpc.vpc01", "tags"),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
						tags = [ "terraform-test", "vpc" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCExists(tt, "scaleway_vpc.vpc01"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_vpc.vpc01", "tags.1", "vpc"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpcAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetVPC(&vpc.GetVPCRequest{
			VpcID:  ID,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc" {
				continue
			}

			vpcAPI, region, ID, err := vpcAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcAPI.GetVPC(&vpc.GetVPCRequest{
				VpcID:  ID,
				Region: region,
			})

			if err == nil {
				return fmt.Errorf("VPC (%s) still exists", rs.Primary.ID)
			}

			if !errs.Is404Error(err) {
				return err
			}
		}

		return nil
	}
}
