### Complete Example with Dependencies

variable "domain_name" {
  type = string
}


resource "scaleway_mnq_sns" "sns" {
}

resource "scaleway_mnq_sns_credentials" "sns_credentials" {
  permissions {
    can_manage = true
  }
}

resource "scaleway_mnq_sns_topic" "sns_topic" {
  name       = "test-mnq-sns-topic-basic"
  access_key = scaleway_mnq_sns_credentials.sns_credentials.access_key
  secret_key = scaleway_mnq_sns_credentials.sns_credentials.secret_key
}

resource "scaleway_tem_domain" "cr01" {
  name       = var.domain_name
  accept_tos = true
}

resource "scaleway_domain_record" "spf" {
  dns_zone = var.domain_name
  type     = "TXT"
  data     = "v=spf1 ${scaleway_tem_domain.cr01.spf_config} -all"
}

resource "scaleway_domain_record" "dkim" {
  dns_zone = var.domain_name
  name     = "${scaleway_tem_domain.cr01.project_id}._domainkey"
  type     = "TXT"
  data     = scaleway_tem_domain.cr01.dkim_config
}

resource "scaleway_domain_record" "mx" {
  dns_zone = var.domain_name
  type     = "MX"
  data     = "."
}

resource "scaleway_domain_record" "dmarc" {
  dns_zone = var.domain_name
  name     = scaleway_tem_domain.cr01.dmarc_name
  type     = "TXT"
  data     = scaleway_tem_domain.cr01.dmarc_config
}

resource "scaleway_tem_domain_validation" "valid" {
  domain_id = scaleway_tem_domain.cr01.id
  region    = scaleway_tem_domain.cr01.region
  timeout   = 3600
}

resource "scaleway_tem_webhook" "webhook" {
  name        = "example-webhook"
  domain_id   = scaleway_tem_domain.cr01.id
  event_types = ["email_delivered", "email_bounced"]
  sns_arn     = scaleway_mnq_sns_topic.sns_topic.arn
  depends_on  = [scaleway_tem_domain_validation.valid, scaleway_mnq_sns_topic.sns_topic]
}
