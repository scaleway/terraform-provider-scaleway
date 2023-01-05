package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
)

func TestAccScalewayMNQCreeds_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQnNamespaceCreedDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  name     = "test-mnq-ns"
					  protocol = "nats"
					}
			
					resource "scaleway_mnq_credential" "main" {
						name = "test-creed-ns"
						namespace_id = scaleway_mnq_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQCreedExists(tt, "scaleway_mnq_credential.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_credential.main", "protocol", "nats"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "nats_credentials.0.content"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "region"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  name     = "test-mnq-sqs"
					  protocol = "sqs_sns"
					}
					
					resource "scaleway_mnq_credential" "main" {
					  name         = "test-creed-sqs"
					  namespace_id = scaleway_mnq_namespace.main.id
					  sqs_sns_credentials {
						permissions {
						  can_publish = true
						  can_receive = true
						  can_manage  = true
						}
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQCreedExists(tt, "scaleway_mnq_credential.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_credential.main", "name", "test-creed-sqs"),
					resource.TestCheckResourceAttr("scaleway_mnq_credential.main", "protocol", "sqs_sns"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_publish"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_manage"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_receive"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.secret_key"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.access_key"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "region"),
				),
			},
			{
				Config: `
					resource "scaleway_mnq_namespace" "main" {
					  name     = "test-mnq-sqs-update"
					  protocol = "sqs_sns"
					}
					
					resource "scaleway_mnq_credential" "main" {
					  name         = "test-mnq-sqs-update"
					  namespace_id = scaleway_mnq_namespace.main.id
					  sqs_sns_credentials {
						permissions {
						  can_publish = true
						  can_receive = true
						  can_manage  = true
						}
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQCreedExists(tt, "scaleway_mnq_credential.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_credential.main", "name", "test-mnq-sqs-update"),
					resource.TestCheckResourceAttr("scaleway_mnq_credential.main", "protocol", "sqs_sns"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_publish"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_manage"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.permissions.0.can_receive"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.secret_key"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_credential.main", "sqs_sns_credentials.0.access_key"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_namespace.main", "region"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQCreedExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetCredential(&mnq.GetCredentialRequest{
			CredentialID: id,
			Region:       region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQnNamespaceCreedDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_credential" {
				continue
			}

			api, region, id, err := mnqAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteCredential(&mnq.DeleteCredentialRequest{
				CredentialID: id,
				Region:       region,
			})

			if err == nil {
				return fmt.Errorf("mnq namespace (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
