resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox" "main" {
  domain_id            = scaleway_mailbox_domain.main.id
  local_part           = "contact"
  password_wo          = var.mailbox_password
  password_wo_version  = 1
  subscription_period  = "monthly"
  forwarding_enabled   = true
  forwarding_keep_copy = true
}

resource "scaleway_mailbox_forwarding" "backup" {
  mailbox_id = scaleway_mailbox.main.id
  email      = "backup@example.com"
}
