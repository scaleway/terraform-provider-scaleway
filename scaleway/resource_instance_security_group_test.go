package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_security_group", &resource.Sweeper{
		Name: "scaleway_instance_security_group",
		F:    testSweepComputeInstanceSecurityGroup,
	})
}

func TestAccScalewayInstanceSecurityGroup_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						name = "sg-name"
						inbound_default_policy = "drop"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "inbound_default_policy", "drop"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "outbound_default_policy", "accept"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						name = "sg-name"
						inbound_default_policy = "accept"
						tags = [ "test-terraform" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "tags.0", "test-terraform"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroup_Policy(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						name = "sg-name"
						description = "terraform-security-group"
						enable_default_security = "false"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "description", "terraform-security-group"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "enable_default_security", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						name = "sg-name"
						description = "terraform-security-group-update"
						enable_default_security = "true"
						tags = [ "test-terraform" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "description", "terraform-security-group-update"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "tags.0", "test-terraform"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "enable_default_security", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "stateful", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						name = "sg-name-update"
						description = "terraform-security-group-update"
						enable_default_security = "true"
						stateful = "false"
						tags = [ "test-terraform" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "name", "sg-name-update"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "description", "terraform-security-group-update"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "tags.0", "test-terraform"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "enable_default_security", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "stateful", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" main {
						name = "sg-on-fr-par-2"
						description = "terraform-security-group"
						enable_default_security = "true"
						tags = [ "test-terraform", "fr-par-2"]
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSecurityGroupExists(tt, "scaleway_instance_security_group.main"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "name", "sg-on-fr-par-2"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "description", "terraform-security-group"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.0", "test-terraform"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "enable_default_security", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "stateful", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "zone", "fr-par-2"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSecurityGroup_Tags(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "main" {
						tags = [ "foo", "bar" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.0", "foo"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.1", "bar"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "main" {
						tags = [ "foo", "buzz" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.0", "foo"),
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.1", "buzz"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "main" {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.main", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceSecurityGroupExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		instanceAPI := instance.NewAPI(tt.Meta.scwClient)
		_, err = instanceAPI.GetSecurityGroup(&instance.GetSecurityGroupRequest{
			SecurityGroupID: ID,
			Zone:            zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstanceSecurityGroupDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceAPI := instance.NewAPI(tt.Meta.scwClient)
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_security_group" {
				continue
			}

			zone, ID, err := parseZonedID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetSecurityGroup(&instance.GetSecurityGroupRequest{
				Zone:            zone,
				SecurityGroupID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("security group (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testSweepComputeInstanceSecurityGroup(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		instanceAPI := instance.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the security groups in (%s)", zone)

		listResp, err := instanceAPI.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing security groups in sweeper: %s", err)
			return nil
		}

		for _, securityGroup := range listResp.SecurityGroups {
			// Can't delete default security group.
			if securityGroup.ProjectDefault {
				continue
			}
			err = instanceAPI.DeleteSecurityGroup(&instance.DeleteSecurityGroupRequest{
				Zone:            zone,
				SecurityGroupID: securityGroup.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting security groups in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayInstanceSecurityGroup_EnableDefaultSecurity(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceSecurityGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						tags = [ "test-terraform" ]
						enable_default_security = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "enable_default_security", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_security_group" "base" {
						tags = [ "test-terraform" ]
						enable_default_security = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_security_group.base", "enable_default_security", "true"),
				),
			},
		},
	})
}
