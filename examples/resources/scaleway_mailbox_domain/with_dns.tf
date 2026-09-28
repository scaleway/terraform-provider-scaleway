variable "domain_name" {
  type = string
}

resource "scaleway_mailbox_domain" "main" {
  name = var.domain_name
}

# Iterate over required DNS records and create them in Scaleway DNS.
resource "scaleway_domain_record" "mailbox_dns" {
  for_each = {
    for rec in scaleway_mailbox_domain.main.dns_records :
    "${rec.dns_type}-${rec.dns_name}" => rec
    if rec.level == "required"
  }

  dns_zone = var.domain_name
  name     = each.value.dns_name
  type     = each.value.dns_type
  data     = each.value.dns_value
}
