data "scaleway_domain_external" "example" {
  domain = "example.com"
}

output "ns_servers" {
  value = data.scaleway_domain_external.example.ns_servers
}
