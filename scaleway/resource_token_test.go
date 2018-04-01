package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_token", &resource.Sweeper{
		Name: "scaleway_token",
		F:    testSweepToken,
	})
}

func testSweepToken(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	scaleway := client.(*Client).scaleway
	log.Printf("[DEBUG] Destroying the tokens in (%s)", region)

	tokens, err := scaleway.GetTokens()
	if err != nil {
		return fmt.Errorf("Error describing tokens in Sweeper: %s", err)
	}

	for _, token := range tokens {
		if err := scaleway.DeleteToken(token.ID); err != nil {
			return fmt.Errorf("Error deleting token in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayToken_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayTokenConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTokenExists("scaleway_token.base"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "description", "just a test"),
					resource.TestCheckResourceAttr(
						"scaleway_token.base", "expires", "false"),
				),
			},
			resource.TestStep{
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayTokenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
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
	client := testAccProvider.Meta().(*Client).scaleway

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

		client := testAccProvider.Meta().(*Client).scaleway
		token, err := client.GetToken(rs.Primary.ID)

		if err != nil {
			return err
		}

		if token.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found")
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
