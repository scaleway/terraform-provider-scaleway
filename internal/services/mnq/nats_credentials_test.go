package mnq_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mnqSDK "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mnq"
	mnqtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mnq/testfuncs"
)

func TestAccNatsCredentials_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectID := mnqtestfuncs.ListProjectID(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isNatsCredentialsDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_mnq_nats_account main {
						project_id = %q
						name = "test-mnq-nats-credentials-basic-test"
					}

					resource scaleway_mnq_nats_credentials main {
						account_id = scaleway_mnq_nats_account.main.id
					}
				`, projectID),
				Check: resource.ComposeTestCheckFunc(
					isNatsCredentialsPresent(tt, "scaleway_mnq_nats_credentials.main"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_nats_credentials.main", "file"),
				),
			},
		},
	})
}

func TestAccNatsCredentials_UpdateName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectID := mnqtestfuncs.ListProjectID(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isNatsCredentialsDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_mnq_nats_account main {
						project_id = %q
						name = "test-mnq-nats-credentials-update"
					}

					resource scaleway_mnq_nats_credentials main {
						account_id = scaleway_mnq_nats_account.main.id
					}
				`, projectID),
				Check: resource.ComposeTestCheckFunc(
					isNatsCredentialsPresent(tt, "scaleway_mnq_nats_credentials.main"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_nats_credentials.main", "file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_mnq_nats_account main {
						project_id = %q
						name = "test-mnq-nats-credentials-update"
					}

					resource scaleway_mnq_nats_credentials main {
						account_id = scaleway_mnq_nats_account.main.id
						name="toto"
					}
				`, projectID),
				Check: resource.ComposeTestCheckFunc(
					isNatsCredentialsPresent(tt, "scaleway_mnq_nats_credentials.main"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_nats_credentials.main", "file"),
				),
			},
		},
	})
}

func isNatsCredentialsPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnq.NewNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), mnqtestfuncs.NamespaceReadRetryTimeout)
		defer cancel()

		return retry.RetryContext(ctx, mnqtestfuncs.NamespaceReadRetryTimeout, func() *retry.RetryError {
			_, err = api.GetNatsCredentials(&mnqSDK.NatsAPIGetNatsCredentialsRequest{
				NatsCredentialsID: id,
				Region:            region,
			})
			if err == nil {
				return nil
			}

			if httperrors.Is404(err) && strings.Contains(err.Error(), "resource namespace") {
				return retry.RetryableError(err)
			}

			if strings.Contains(err.Error(), "insufficient permissions: read namespace") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		})
	}
}

func isNatsCredentialsDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_nats_credentials" {
				continue
			}

			api, region, id, err := mnq.NewNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteNatsCredentials(&mnqSDK.NatsAPIDeleteNatsCredentialsRequest{
				NatsCredentialsID: id,
				Region:            region,
			})
			if err == nil {
				return fmt.Errorf("mnq nats credentials (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
