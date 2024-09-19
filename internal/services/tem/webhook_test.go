package tem_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

const (
	webhookName       = "terraform-webhook-test"
	organizationID    = "105bdce1-64c0-48ab-899d-868455867ecf"
	webhookDomainName = "webhook-test.scaleway-terraform.com"
)

func TestAccWebhook_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	eventTypes := []string{"email_delivered", "email_dropped"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isWebhookDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "project" {
						name = "default"
						organization_id      = "%s"
					}

					data scaleway_mnq_sns sns {
						project_id= data.scaleway_account_project.project.project_id
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = data.scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = data.scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					data "scaleway_tem_domain" "cr01" {
						name = "%s"
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = data.scaleway_tem_domain.cr01.id
						event_types = ["%s", "%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_mnq_sns_topic.sns_topic]
					}
				`, organizationID, webhookDomainName, webhookName, eventTypes[0], eventTypes[1]),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", webhookName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "2"),
				),
			},
		},
	})
}

func TestAccWebhook_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	initialName := "terraform-webhook-test"
	updatedName := "terraform-webhook-updated"
	eventTypes := []string{"email_delivered"}
	updatedEventTypes := []string{"email_queued"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isWebhookDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "project" {
						name = "default"
						organization_id      = "%s"
					}

					data scaleway_mnq_sns sns {
						project_id= data.scaleway_account_project.project.project_id
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = data.scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = data.scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					data "scaleway_tem_domain" "cr01" {
						name = "%s"
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = data.scaleway_tem_domain.cr01.id
						event_types = ["%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_mnq_sns_topic.sns_topic]
					}
				`, organizationID, webhookDomainName, initialName, eventTypes[0]),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", initialName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "project" {
						name = "default"
						organization_id      = "%s"
					}

					data scaleway_mnq_sns sns {
						project_id= data.scaleway_account_project.project.project_id
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = data.scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = data.scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					data "scaleway_tem_domain" "cr01" {
						name = "%s"
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = data.scaleway_tem_domain.cr01.id
						event_types = ["%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_mnq_sns_topic.sns_topic]
					}
				`, organizationID, webhookDomainName, updatedName, updatedEventTypes[0]),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", updatedName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "1"),
				),
			},
		},
	})
}

func isWebhookPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
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
			return err
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
			errorCode := httperrors.Is404(err)
			_ = errorCode
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
