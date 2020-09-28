package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalewayToken_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTokenExists("scaleway_token.base"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "description", "just a test"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "expires", "false"),
					resource.TestCheckResourceAttrSet("scaleway_token.base", "access_key"),
					resource.TestCheckResourceAttrSet("scaleway_token.base", "secret_key"),
					resource.TestCheckResourceAttrSet("scaleway_token.base", "creation_ip"),
				),
			},
			{
				Config: testAccCheckScalewayTokenConfig_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTokenExists("scaleway_token.base"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "description", "just a test 222"),
					resource.TestCheckResourceAttrSet("scaleway_token.base", "expiration_date"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "expires", "true"),
				),
			},
		},
	})
}

func TestAccScalewayToken_Expiry(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayTokenConfig_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTokenExists("scaleway_token.base"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "description", "just a test 222"),
					resource.TestCheckResourceAttrSet("scaleway_token.base", "expiration_date"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "expires", "true"),
				),
			},
		},
	})
}

func testAccCheckScalewayTokenDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Meta).deprecatedClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetToken(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Token still exists")
		}
	}

	return nil
}

func testAccCheckScalewayTokenExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Token ID is set")
		}

		client := testAccProvider.Meta().(*Meta).deprecatedClient
		token, err := client.GetToken(rs.Primary.ID)

		if err != nil {
			return err
		}

		if token.ID != rs.Primary.ID {
			return fmt.Errorf("Expected %q, got %q", rs.Primary.ID, token.ID)
		}

		return nil
	}
}

var testAccCheckScalewayTokenConfig = `
resource "scaleway_token" "base" {
	expires = false
	description = "just a test"
}`

var testAccCheckScalewayTokenConfig_Update = `
resource "scaleway_token" "base" {
	expires = true
	description = "just a test 222"
}`
