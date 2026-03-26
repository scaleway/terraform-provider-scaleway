resource "scaleway_domain_external" "example" {
  domain = "scw.example.com"
}

resource "scaleway_domain_record" "validation" {
  dns_zone = "example.com"
  name     = "_scaleway-challenge.scw"
  type     = "TXT"
  data     = scaleway_domain_external.example.validation_token
}

# Wait for domain validation (default timeout is 30 minutes)
resource "scaleway_domain_external_validated" "example" {
  domain     = scaleway_domain_external.example.domain
  depends_on = [scaleway_domain_record.validation]
}

# After validation, we can use the domain's NS servers
resource "scaleway_domain_record" "ns" {
  for_each = {
    ns0 = scaleway_domain_external_validated.example.ns_servers[0]
    ns1 = scaleway_domain_external_validated.example.ns_servers[1]
  }
  dns_zone = "example.com"
  name     = "scw"
  type     = "NS"
  data     = each.value
}