resource "scaleway_mailbox_domain" "corp" {
  name = "corp.example.com"
}

locals {
  mailboxes = toset(["alice", "bob"])
}

resource "scaleway_mailbox_mailbox" "employees" {
  for_each = local.mailboxes

  domain_id           = scaleway_mailbox_domain.corp.id
  local_part          = each.key
  password            = var.mailbox_password
  subscription_period = "monthly"
}
