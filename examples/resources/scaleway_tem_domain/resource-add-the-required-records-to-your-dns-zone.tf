### Add the required records to your DNS zone

variable "domain_name" {
  type = string
}

resource "scaleway_tem_domain" "main" {
  name       = var.domain_name
  accept_tos = true
}

resource "scaleway_domain_record" "spf" {
  dns_zone = var.domain_name
  type     = "TXT"
  data     = scaleway_tem_domain.main.spf_value
}

resource "scaleway_domain_record" "dkim" {
  dns_zone = var.domain_name
  name     = scaleway_tem_domain.main.dkim_name
  type     = "TXT"
  data     = scaleway_tem_domain.main.dkim_config
}

resource "scaleway_domain_record" "mx" {
  dns_zone = var.domain_name
  type     = "MX"
  data     = scaleway_tem_domain.main.mx_config
}

resource "scaleway_domain_record" "dmarc" {
  dns_zone = var.domain_name
  name     = scaleway_tem_domain.main.dmarc_name
  type     = "TXT"
  data     = scaleway_tem_domain.main.dmarc_config
}
