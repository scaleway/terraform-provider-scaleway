resource "scaleway_domain_external" "example" {
  domain = "example.com"
}

output "validation_token" {
  value = scaleway_domain_external.example.validation_token
}