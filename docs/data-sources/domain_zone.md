---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_zone"
---

# scaleway_domain_zone

The `scaleway_domain_record` data source is used to get information about a DNS zone within a specific domain and subdomain in Scaleway Domains and DNS.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.

## Query a domain zone

The following command allows you to retrieve information about the DNS zone for the subdomain `test` within the domain `scaleway-terraform.com`.

```hcl
# Get zone
data "scaleway_domain_zone" "main" {
  domain    = "scaleway-terraform.com"
  subdomain = "test"
}
```

## Arguments reference

This section lists the arguments that can be provided to the `scaleway_domain_record` data source:


- `domain` - (Required) The primary domain name where the DNS zone is located. This is a mandatory field.

- `subdomain` - (Required) The subdomain (or zone name) within the primary domain. This is a mandatory field.

- `project_id` - ([Defaults to the provider's](../index.md#project_id) `project_id`). The ID of the Scaleway Project associated with the domain. If not specified, it defaults to the `project_id` set in the provider configuration.

## Attributes reference

This section lists the attributes that are exported by the `scaleway_domain_zone` data source. These attributes can be referenced in other parts of your Terraform configuration:

- `id` - The unique identifier of the zone, formatted as `{subdomain}.{domain}`.

- `ns` - The list of name servers for the zone.

- `ns_default` - The default list of name servers for the zone.

- `ns_master` - The master list of name servers for the zone.

- `status` - The status of the domain zone.

- `message` - Message associated with the domain zone (typically used for status or error messages).

- `updated_at` - The date and time of the last update to the DNS zone.
