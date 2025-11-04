package tem_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

// TestAccBlockedList_Basic is now covered by TestAccTEM_Complete step 2

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
