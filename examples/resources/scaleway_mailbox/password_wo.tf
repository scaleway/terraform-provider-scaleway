resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox" "support" {
  domain_id           = scaleway_mailbox_domain.main.id
  local_part          = "support"
  password_wo         = var.mailbox_password
  password_wo_version = 1
  subscription_period = "monthly"
}
