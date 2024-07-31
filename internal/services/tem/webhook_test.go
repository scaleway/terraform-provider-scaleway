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

const webhookName = "terraform-webhook-test"

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
					

					resource "scaleway_mnq_sns" "sns" {
					}
 
					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-basic"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}

					resource "scaleway_domain_record" "spf" {
  						dns_zone = "%s"
  						type     = "TXT"
						data     = "v=spf1 ${scaleway_tem_domain.cr01.spf_config} -all"
					}

					resource "scaleway_domain_record" "dkim" {
  						dns_zone = "%s"
  						name     = "${scaleway_tem_domain.cr01.project_id}._domainkey"
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dkim_config
					}
					resource "scaleway_domain_record" "mx" {
  						dns_zone = "%s"
  						type     = "MX"
  						data     = "."
					}

					resource "scaleway_domain_record" "dmarc" {
						dns_zone = "%s"
  						name     = scaleway_tem_domain.cr01.dmarc_name
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dmarc_config
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 3600
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = scaleway_tem_domain.cr01.id
						event_types = ["%s", "%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_tem_domain_validation.valid, scaleway_mnq_sns_topic.sns_topic]
					}
				`, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, webhookName, eventTypes[0], eventTypes[1]),
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

					resource "scaleway_mnq_sns" "sns" {
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-update"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}

					resource "scaleway_domain_record" "spf" {
  						dns_zone = "%s"
  						type     = "TXT"
						data     = "v=spf1 ${scaleway_tem_domain.cr01.spf_config} -all"
					}

					resource "scaleway_domain_record" "dkim" {
  						dns_zone = "%s"
  						name     = "${scaleway_tem_domain.cr01.project_id}._domainkey"
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dkim_config
					}
					resource "scaleway_domain_record" "mx" {
  						dns_zone = "%s"
  						type     = "MX"
  						data     = "."
					}

					resource "scaleway_domain_record" "dmarc" {
						dns_zone = "%s"
  						name     = scaleway_tem_domain.cr01.dmarc_name
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dmarc_config
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 3600
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = scaleway_tem_domain.cr01.id
						event_types = ["%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_tem_domain_validation.valid, scaleway_mnq_sns_topic.sns_topic]
					}
				`, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, initialName, eventTypes[0]),
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

					resource "scaleway_mnq_sns" "sns" {
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials"  {
						project_id = scaleway_mnq_sns.sns.project_id
						permissions {
							can_manage = true
						}
					}

					resource "scaleway_mnq_sns_topic" "sns_topic" {
						project_id = scaleway_mnq_sns.sns.project_id
						name = "test-mnq-sns-topic-update"
						access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
						secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
					}

					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}

					resource "scaleway_domain_record" "spf" {
  						dns_zone = "%s"
  						type     = "TXT"
						data     = "v=spf1 ${scaleway_tem_domain.cr01.spf_config} -all"
					}

					resource "scaleway_domain_record" "dkim" {
  						dns_zone = "%s"
  						name     = "${scaleway_tem_domain.cr01.project_id}._domainkey"
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dkim_config
					}
					resource "scaleway_domain_record" "mx" {
  						dns_zone = "%s"
  						type     = "MX"
  						data     = "."
					}

					resource "scaleway_domain_record" "dmarc" {
						dns_zone = "%s"
  						name     = scaleway_tem_domain.cr01.dmarc_name
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dmarc_config
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 3600
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = scaleway_tem_domain.cr01.id
						event_types = ["%s"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on = [scaleway_tem_domain_validation.valid, scaleway_mnq_sns_topic.sns_topic]
					}
				`, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, updatedName, updatedEventTypes[0]),
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
