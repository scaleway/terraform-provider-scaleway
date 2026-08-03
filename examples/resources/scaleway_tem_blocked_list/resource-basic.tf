### Basic

resource "scaleway_tem_blocked_list" "test" {
  domain_id = "fr-par/12345678-1234-1234-1234-123456789abc"
  email     = "spam@example.com"
  type      = "mailbox_full"
  reason    = "Spam detected"
  region    = "fr-par"
}
