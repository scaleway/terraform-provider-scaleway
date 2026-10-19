package tem_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const testAccTEMProjectID = "105bdce1-64c0-48ab-899d-868455867ecf"

func temCompleteBaseConfig(subDomainName string) string {
	return fmt.Sprintf(`
		resource "scaleway_domain_zone" "test" {
			domain     = "%s"
			subdomain  = "%s"
			project_id = "%s"
		}

		resource scaleway_tem_domain cr01 {
			name       = scaleway_domain_zone.test.id
			accept_tos = true
			autoconfig = true
			project_id = "%s"
		}

		resource scaleway_tem_domain_validation valid {
			domain_id = scaleway_tem_domain.cr01.id
			region    = scaleway_tem_domain.cr01.region
			timeout   = 3600
		}
	`, domainNameValidation, subDomainName, testAccTEMProjectID, testAccTEMProjectID)
}

func temWebhookConfig(webhookName string, eventTypes string) string {
	return fmt.Sprintf(`
		resource "scaleway_mnq_sns" "sns" {
			project_id = "%s"
		}

		resource "scaleway_mnq_sns_credentials" "sns_credentials" {
			project_id = scaleway_mnq_sns.sns.project_id
			permissions {
				can_manage = true
			}
		}

		resource "scaleway_mnq_sns_topic" "sns_topic" {
			project_id = scaleway_mnq_sns.sns.project_id
			name       = "tf-test-tem-sns-topic"
			access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
			secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
			depends_on = [scaleway_mnq_sns_credentials.sns_credentials]
		}

		resource "scaleway_tem_webhook" "webhook" {
			name        = "%s"
			domain_id   = scaleway_tem_domain.cr01.id
			event_types = %s
			sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
			depends_on  = [
				scaleway_mnq_sns_topic.sns_topic,
				scaleway_tem_domain_validation.valid
			]
		}
	`, testAccTEMProjectID, webhookName, eventTypes)
}

func TestAccTEM_Complete(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "tf-test-tem-complete"
	blockedEmail := "spam@example.com"
	webhookName := "tf-test-tem-webhook"
	updatedWebhookName := "tf-test-tem-webhook-updated"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkTEMResourcesDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: temCompleteBaseConfig(subDomainName),
				Check: resource.ComposeTestCheckFunc(
					isDomainPresent(tt, "scaleway_tem_domain.cr01"),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "name", subDomainName+"."+domainNameValidation),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "autoconfig", "true"),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "dmarc_config", "v=DMARC1; p=none"),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "dmarc_name", regexp.MustCompile(`^_dmarc\.`+regexp.QuoteMeta(subDomainName+"."+domainNameValidation)+`\.$`)),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "dkim_name", regexp.MustCompile(`^[a-f0-9-]+\._domainkey\.`+regexp.QuoteMeta(subDomainName+"."+domainNameValidation)+`\.$`)),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "spf_value", regexp.MustCompile(`^v=spf1 include:.+ -all$`)),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "mx_config", "10 blackhole.tem.scaleway.com."),
					acctest.CheckResourceAttrUUID("scaleway_tem_domain.cr01", "id"),
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "true"),
				),
			},
			{
				ResourceName:            "scaleway_tem_domain.cr01",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status", "last_valid_at", "next_check_at", "reputation"},
			},
			{
				Config: temCompleteBaseConfig(subDomainName) + fmt.Sprintf(`
					resource "scaleway_tem_blocked_list" "test" {
						domain_id  = scaleway_tem_domain.cr01.id
						email      = "%s"
						type       = "mailbox_full"
						reason     = "Spam detected"
						region     = "fr-par"
						depends_on = [scaleway_tem_domain_validation.valid]
					}
				`, blockedEmail),
				Check: resource.ComposeTestCheckFunc(
					isBlockedEmailPresent(tt, "scaleway_tem_blocked_list.test"),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "email", blockedEmail),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "type", "mailbox_full"),
				),
			},
			{
				Config: temCompleteBaseConfig(subDomainName) + temWebhookConfig(webhookName, `["email_delivered", "email_dropped"]`),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", webhookName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "2"),
				),
			},
			{
				Config: temCompleteBaseConfig(subDomainName) + temWebhookConfig(updatedWebhookName, `["email_queued"]`),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", updatedWebhookName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "1"),
				),
			},
			{
				Config: temCompleteBaseConfig(subDomainName) + `
					data "scaleway_tem_domain" "test" {
						name = scaleway_tem_domain.cr01.name
						depends_on = [scaleway_tem_domain_validation.valid]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_tem_domain.test", "reputation.0.status"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_domain.test", "reputation.0.score"),
				),
			},
		},
	})
}

func checkTEMResourcesDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if err := isDomainDestroyed(tt)(state); err != nil {
			return err
		}

		return isWebhookDestroyed(tt)(state)
	}
}
