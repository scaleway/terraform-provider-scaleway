### Query domain records

# Query record by DNS zone, record name, type and content
data "scaleway_domain_record" "by_content" {
  dns_zone = "domain.tld"
  name     = "www"
  type     = "A"
  data     = "1.2.3.4"
}

# Query record by DNS zone and record ID
data "scaleway_domain_record" "by_id" {
  dns_zone  = "domain.tld"
  record_id = "11111111-1111-1111-1111-111111111111"
}
