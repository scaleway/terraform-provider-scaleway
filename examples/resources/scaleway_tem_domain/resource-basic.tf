### Basic

resource "scaleway_tem_domain" "main" {
  accept_tos = true
  name       = "example.com"
}
