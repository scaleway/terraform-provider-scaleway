package mnq_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mnq"
)

func TestAccSNSTopicSubscription_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isSNSTopicSubscriptionDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sns_topic_subscription_basic"
					}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						permissions {
							can_manage = true
							can_publish = true
							can_receive = true
						}
					}

					resource scaleway_mnq_sns_topic main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.main.access_key
						secret_key = scaleway_mnq_sns_credentials.main.secret_key
					}
					
					resource scaleway_mnq_sns_topic_subscription by_id {
						project_id = scaleway_mnq_sns.main.project_id
						access_key = scaleway_mnq_sns_credentials.main.access_key
						secret_key = scaleway_mnq_sns_credentials.main.secret_key
						topic_id = scaleway_mnq_sns_topic.main.id
						protocol = "http"
						endpoint = "http://scaleway.com"
					}

					resource scaleway_mnq_sns_topic_subscription by_arn {
						project_id = scaleway_mnq_sns.main.project_id
						access_key = scaleway_mnq_sns_credentials.main.access_key
						secret_key = scaleway_mnq_sns_credentials.main.secret_key
						topic_arn = scaleway_mnq_sns_topic.main.arn
						protocol = "http"
						endpoint = "http://scaleway.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isSNSTopicSubscriptionPresent(tt, "scaleway_mnq_sns_topic_subscription.by_id"),
					acctest.CheckResourceAttrUUID("scaleway_mnq_sns_topic_subscription.by_id", "id"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic_subscription.by_id", "protocol", "http"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic_subscription.by_id", "endpoint", "http://scaleway.com"),

					isSNSTopicSubscriptionPresent(tt, "scaleway_mnq_sns_topic_subscription.by_arn"),
					acctest.CheckResourceAttrUUID("scaleway_mnq_sns_topic_subscription.by_arn", "id"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic_subscription.by_arn", "protocol", "http"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_topic_subscription.by_arn", "endpoint", "http://scaleway.com"),
				),
			},
		},
	})
}

func isSNSTopicSubscriptionPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		arn, err := mnq.DecomposeMNQSubscriptionID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}

		snsClient, err := mnq.NewSNSClient(tt.Meta.HTTPClient(), arn.Region.String(), rs.Primary.Attributes["sns_endpoint"], rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"])
		if err != nil {
			return err
		}

		_, err = snsClient.GetSubscriptionAttributes(&sns.GetSubscriptionAttributesInput{
			SubscriptionArn: scw.StringPtr(arn.String()),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isSNSTopicSubscriptionDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_sns_topic_subscription" {
				continue
			}

			arn, err := mnq.DecomposeMNQSubscriptionID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("failed to parse id: %w", err)
			}

			snsClient, err := mnq.NewSNSClient(tt.Meta.HTTPClient(), arn.Region.String(), rs.Primary.Attributes["sns_endpoint"], rs.Primary.Attributes["access_key"], rs.Primary.Attributes["secret_key"])
			if err != nil {
				return err
			}

			_, err = snsClient.GetSubscriptionAttributes(&sns.GetSubscriptionAttributesInput{
				SubscriptionArn: scw.StringPtr(arn.String()),
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
					return nil
				}
				return err
			}

			return fmt.Errorf("mnq sns subscription (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}
