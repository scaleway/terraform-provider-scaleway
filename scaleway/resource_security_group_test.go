package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_security_group", &resource.Sweeper{
		Name: "scaleway_security_group",
		F:    testSweepSecurityGroup,
	})
}

func testSweepSecurityGroup(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	scaleway := client.(*Client).scaleway
	log.Printf("[DEBUG] Destroying the security groups in (%s)", region)

	sgs, err := scaleway.GetSecurityGroups()
	if err != nil {
		return fmt.Errorf("Error describing security groups in Sweeper: %s", err)
	}

	for _, sg := range sgs {
		if sg.OrganizationDefault {
			continue
		}

		if err := scaleway.DeleteSecurityGroup(sg.ID); err != nil {
			return fmt.Errorf("Error deleting ip in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewaySecurityGroup_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewaySecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewaySecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewaySecurityGroupExists("scaleway_security_group.base"),
					testAccCheckScalewaySecurityGroupAttributes("scaleway_security_group.base"),
					resource.TestCheckResourceAttr("scaleway_security_group.base", "name", "public"),
					resource.TestCheckResourceAttr("scaleway_security_group.base", "description", "public gateway"),
				),
			},
		},
	})
}

func testAccCheckScalewaySecurityGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scaleway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetSecurityGroup(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Security Group still exists")
		}
	}

	return nil
}

func testAccCheckScalewaySecurityGroupAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Client).scaleway
		group, err := client.GetSecurityGroup(rs.Primary.ID)
		if err != nil {
			return err
		}

		if group.Name != "public" {
			return fmt.Errorf("Security Group has wrong name")
		}
		if group.Description != "public gateway" {
			return fmt.Errorf("Security Group has wrong description")
		}

		return nil
	}
}

func testAccCheckScalewaySecurityGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group ID is set")
		}

		client := testAccProvider.Meta().(*Client).scaleway
		group, err := client.GetSecurityGroup(rs.Primary.ID)

		if err != nil {
			return err
		}

		if group.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

var testAccCheckScalewaySecurityGroupConfig = `
resource "scaleway_security_group" "base" {
  name = "public"
  description = "public gateway"
}
`
