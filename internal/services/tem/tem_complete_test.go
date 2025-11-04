package tem_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestAccTEM_Complete tests multiple TEM resources using a single validated domain to avoid quota issues
// Step 1: Domain validation with autoconfig
// Step 2: Blockedlist
// Step 3: Webhook creation
// Step 4: Webhook update (name and event_types)
// Step 5: Data source reputation
func TestAccTEM_Complete(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "test-complete"
	blockedEmail := "spam@example.com"
	webhookName := "terraform-webhook-test"
	updatedWebhookName := "terraform-webhook-updated"
	organizationID := "105bdce1-64c0-48ab-899d-868455867ecf"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkTEMResourcesDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
						domain_id = scaleway_tem_domain.cr01.id
						region    = scaleway_tem_domain.cr01.region
						timeout   = 3600
					}
				`, domainNameValidation, subDomainName),
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
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
						domain_id = scaleway_tem_domain.cr01.id
						region    = scaleway_tem_domain.cr01.region
						timeout   = 3600
					}

					resource "scaleway_tem_blocked_list" "test" {
						domain_id  = scaleway_tem_domain.cr01.id
						email      = "%s"
						type       = "mailbox_full"
						reason     = "Spam detected"
						region     = "fr-par"
						depends_on = [scaleway_tem_domain_validation.valid]
					}
				`, domainNameValidation, subDomainName, blockedEmail),
				Check: resource.ComposeTestCheckFunc(
					isBlockedEmailPresent(tt, "scaleway_tem_blocked_list.test"),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "email", blockedEmail),
					resource.TestCheckResourceAttr("scaleway_tem_blocked_list.test", "type", "mailbox_full"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
						domain_id = scaleway_tem_domain.cr01.id
						region    = scaleway_tem_domain.cr01.region
						timeout   = 3600
					}

					data scaleway_account_project "project" {
						name = "default"
						organization_id = "%s"
					}

					data scaleway_mnq_sns sns {
						project_id = data.scaleway_account_project.project.project_id
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials" {
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
						depends_on = [scaleway_mnq_sns_credentials.sns_credentials]
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = scaleway_tem_domain.cr01.id
						event_types = ["email_delivered", "email_dropped"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on  = [
							scaleway_mnq_sns_topic.sns_topic,
							scaleway_tem_domain_validation.valid
						]
					}
				`, domainNameValidation, subDomainName, organizationID, webhookName),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", webhookName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "2"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
						domain_id = scaleway_tem_domain.cr01.id
						region    = scaleway_tem_domain.cr01.region
						timeout   = 3600
					}

					data scaleway_account_project "project" {
						name = "default"
						organization_id = "%s"
					}

					data scaleway_mnq_sns sns {
						project_id = data.scaleway_account_project.project.project_id
					}

					resource "scaleway_mnq_sns_credentials" "sns_credentials" {
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
						depends_on = [scaleway_mnq_sns_credentials.sns_credentials]
					}

					resource "scaleway_tem_webhook" "webhook" {
						name        = "%s"
						domain_id   = scaleway_tem_domain.cr01.id
						event_types = ["email_queued"]
						sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
						depends_on  = [
							scaleway_mnq_sns_topic.sns_topic,
							scaleway_tem_domain_validation.valid
						]
					}
				`, domainNameValidation, subDomainName, organizationID, updatedWebhookName),
				Check: resource.ComposeTestCheckFunc(
					isWebhookPresent(tt, "scaleway_tem_webhook.webhook"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "name", updatedWebhookName),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "domain_id"),
					resource.TestCheckResourceAttrSet("scaleway_tem_webhook.webhook", "sns_arn"),
					resource.TestCheckResourceAttr("scaleway_tem_webhook.webhook", "event_types.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
						domain_id = scaleway_tem_domain.cr01.id
						region    = scaleway_tem_domain.cr01.region
						timeout   = 3600
					}

					data "scaleway_tem_domain" "test" {
						name = scaleway_tem_domain.cr01.name
						depends_on = [scaleway_tem_domain_validation.valid]
					}
				`, domainNameValidation, subDomainName),
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

		if err := isWebhookDestroyed(tt)(state); err != nil {
			return err
		}

		return nil
	}
}
