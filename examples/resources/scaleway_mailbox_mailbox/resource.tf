resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox_mailbox" "john" {
  domain_id           = scaleway_mailbox_domain.main.id
  local_part          = "john.doe"
  password            = var.mailbox_password
  subscription_period = "monthly"
}
