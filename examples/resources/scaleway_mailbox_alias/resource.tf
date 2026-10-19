resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox" "john" {
  domain_id           = scaleway_mailbox_domain.main.id
  local_part          = "john.doe"
  password_wo         = var.mailbox_password
  password_wo_version = 1
  subscription_period = "monthly"
}

resource "scaleway_mailbox_alias" "contact" {
  mailbox_id  = scaleway_mailbox.john.id
  local_part  = "contact"
  description = "Public contact address"
}
