---
page_title: "Scaleway: scaleway_domain_record"
subcategory: "Domains and DNS"
description: |-
  Lists Scaleway DNS zone records across projects and DNS zones.
---

# Resource: scaleway_domain_record

Lists Scaleway DNS zone records across projects and DNS zones.

For more information, see the [product documentation](https://www.scaleway.com/en/docs/domains-and-dns/).


## Example Usage

```terraform
// List DNS zone records across all zones in a project
list "scaleway_domain_record" "all_zones" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    dns_zones   = ["*"]
  }
}
```

```terraform
// List DNS zone records in a specific zone
list "scaleway_domain_record" "by_dns_zone" {
  provider = scaleway

  config {
    dns_zones = ["www.example.com"]
    type      = "A"
  }
}
```

```terraform
// List DNS zone records filtered by name
list "scaleway_domain_record" "by_name" {
  provider = scaleway

  config {
    dns_zones = ["www.example.com"]
    name      = "www"
    type      = "A"
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If omitted, the provider default project is used.
- `dns_zones` - (Required) DNS zone FQDNs to list records from. Use `["*"]` or `["all"]` to list records across all zones in each selected project. Must contain at least one value.
- `name` - (Optional) Name of the DNS zone record to filter on.
- `type` - (Optional) Type of the DNS zone record to filter on.

## Attributes Reference

Each result corresponds to one DNS zone record and exposes the same attributes as the [`scaleway_domain_record` resource](../resources/domain_record.md).
