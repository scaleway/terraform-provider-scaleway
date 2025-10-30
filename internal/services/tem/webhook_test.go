package tem_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

// TestAccWebhook_BasicAndUpdate is now covered by TestAccTEM_Complete steps 3-4

func isWebhookPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", n)
		}

		api, region, id, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetWebhook(&temSDK.GetWebhookRequest{
			WebhookID: id,
			Region:    region,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("webhook not present: %w", err)
		}

		return nil
	}
}

func isWebhookDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_tem_webhook" {
				continue
			}

			api, region, id, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetWebhook(&temSDK.GetWebhookRequest{
				WebhookID: id,
				Region:    region,
			}, scw.WithContext(context.Background()))
			if err == nil {
				return fmt.Errorf("webhook still exists: %s", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return fmt.Errorf("unexpected error on GetWebhook after destroy: %w", err)
			}
		}

		return nil
	}
}
