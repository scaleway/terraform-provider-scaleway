### Basic

resource "scaleway_tem_webhook" "main" {
  domain_id   = "your-domain-id"
  event_types = ["email_delivered", "email_bounced"]
  sns_arn     = "arn:scw:sns:fr-par:project-xxxx:your-sns-topic"
  name        = "example-webhook"
}
