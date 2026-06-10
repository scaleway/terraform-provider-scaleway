---
page_title: "Scaleway: scaleway_domain_zone"
subcategory: "Domains and DNS"
description: |-
  Lists Scaleway DNS zones across projects and domains.
---

# Resource: scaleway_domain_zone

For more information, see the [product documentation](https://www.scaleway.com/en/docs/domains-and-dns/).

## Example Usage

```terraform
// List DNS zones across all domains in a project
list "scaleway_domain_zone" "all_domains" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    domains     = ["*"]
  }
}
```
```terraform
// List a specific DNS zone by FQDN
list "scaleway_domain_zone" "by_dns_zone" {
  provider = scaleway

  config {
    domains   = ["example.com"]
    dns_zones = ["www.example.com"]
  }
}
```
```terraform
// List DNS zones for a domain in the default project
list "scaleway_domain_zone" "by_domain" {
  provider = scaleway

  config {
    domains = ["example.com"]
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If omitted, the provider default project is used.
- `domains` - (Required) Domain apex names to list DNS zones for. Use `["*"]` to list zones across all domains. Must contain at least one value.
- `dns_zones` - (Optional) Filter by DNS zone FQDNs (for example `subdomain.example.com` or `example.com` for a root zone).
- `created_after` - (Optional) Only list DNS zones created after this date (RFC3339).
- `created_before` - (Optional) Only list DNS zones created before this date (RFC3339).
- `updated_after` - (Optional) Only list DNS zones updated after this date (RFC3339).
- `updated_before` - (Optional) Only list DNS zones updated before this date (RFC3339).

## Attributes Reference

Each result corresponds to one DNS zone and exposes the same attributes as the [`scaleway_domain_zone` resource](../resources/domain_zone.md).
