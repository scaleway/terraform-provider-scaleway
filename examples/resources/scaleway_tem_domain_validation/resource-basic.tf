### Basic

resource "scaleway_tem_domain" "main" {
  accept_tos = true
  name       = "example.com"
}

resource "scaleway_tem_domain_validation" "example" {
  domain_id = scaleway_tem_domain.main.id
  region    = "fr-par"
  timeout   = 300
}
