package acctest

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// TestAccCheckScalewayResourceIDChanged checks that the ID of the resource has indeed changed, in case of ForceNew for example.
// It will fail if resourceID is empty so be sure to use acctest.TestAccCheckScalewayResourceIDPersisted first in a test suite.
func TestAccCheckScalewayResourceIDChanged(resourceName string, resourceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if resourceID == nil || *resourceID == "" {
			return errors.New("resourceID was not set")
		}
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource was not found: %s", resourceName)
		}
		if *resourceID == rs.Primary.ID {
			return errors.New("resource ID persisted when it should have changed")
		}
		*resourceID = rs.Primary.ID
		return nil
	}
}

// TestAccCheckScalewayResourceIDPersisted checks that the ID of the resource is the same throughout tests of migration or mutation
// It can be used to check that no ForceNew has been done
func TestAccCheckScalewayResourceIDPersisted(resourceName string, resourceID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource was not found: %s", resourceName)
		}
		if *resourceID != "" && *resourceID != rs.Primary.ID {
			return errors.New("resource ID changed when it should have persisted")
		}
		*resourceID = rs.Primary.ID
		return nil
	}
}

// TestAccCheckScalewayResourceRawIDMatches asserts the equality of IDs from two specified attributes of two Scaleway resources.
func TestAccCheckScalewayResourceRawIDMatches(res1, attr1, res2, attr2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs1, ok1 := s.RootModule().Resources[res1]
		if !ok1 {
			return fmt.Errorf("not found: %s", res1)
		}

		rs2, ok2 := s.RootModule().Resources[res2]
		if !ok2 {
			return fmt.Errorf("not found: %s", res2)
		}

		id1 := locality.ExpandID(rs1.Primary.Attributes[attr1])
		id2 := locality.ExpandID(rs2.Primary.Attributes[attr2])

		if id1 != id2 {
			return fmt.Errorf("ID mismatch: %s from resource %s does not match ID %s from resource %s", id1, res1, id2, res2)
		}

		return nil
	}
}
