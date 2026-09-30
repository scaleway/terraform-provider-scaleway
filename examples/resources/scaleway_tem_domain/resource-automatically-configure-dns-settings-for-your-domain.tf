### Automatically Configure DNS Settings for Your Domain

variable "domain_name" {
  type = string
}

resource "scaleway_tem_domain" "main" {
  name       = var.domain_name
  accept_tos = true
  autoconfig = true
}
