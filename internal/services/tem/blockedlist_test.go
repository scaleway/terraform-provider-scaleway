package tem_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

func TestAccBlockedList_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "test-blockedlist"

	blockedEmail := "spam@example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isBlockedEmailDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: getValidatedDomainConfig(subDomainName) + fmt.Sprintf(`
					resource "scaleway_tem_blocked_list" "test" {
						domain_id  = scaleway_tem_domain.cr01.id
						email      = "%s"
						type       = "mailbox_full"
						reason     = "Spam detected"
						region     = "fr-par"
 						depends_on = [
    						scaleway_tem_domain_validation.valid
  						]
					}
				`, blockedEmail),
				Check: resource.ComposeTestCheckFunc(
					isBlockedEmailPresent(tt, "scaleway_tem_blocked_list.test"),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "email", blockedEmail),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "type", "mailbox_full"),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "reason", "Spam detected"),
					acctest.CheckResourceAttrUUID("scaleway_tem_blocked_list.test", "id"),
				),
			},
		},
	})
}

func isBlockedEmailPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, domainID, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.Attributes["domain_id"])
		if err != nil {
			return err
		}

		blockedEmail := rs.Primary.Attributes["email"]

		blocklists, err := api.ListBlocklists(&temSDK.ListBlocklistsRequest{
			Region:   region,
			DomainID: domainID,
			Email:    scw.StringPtr(blockedEmail),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		if len(blocklists.Blocklists) == 0 {
			return fmt.Errorf("blocked email %s not found in blocklist", blockedEmail)
		}

		return nil
	}
}

func isBlockedEmailDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_tem_blocked_list" {
				continue
			}

			api, region, domainID, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.Attributes["domain_id"])
			if err != nil {
				return err
			}

			blockedEmail := rs.Primary.Attributes["email"]

			blocklists, err := api.ListBlocklists(&temSDK.ListBlocklistsRequest{
				Region:   region,
				DomainID: domainID,
				Email:    scw.StringPtr(blockedEmail),
			}, scw.WithContext(context.Background()))
			if err != nil {
				return err
			}

			if len(blocklists.Blocklists) > 0 {
				return fmt.Errorf("blocked email %s still present after deletion", blockedEmail)
			}
		}

		return nil
	}
}
