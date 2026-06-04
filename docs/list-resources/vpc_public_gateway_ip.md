---
page_title: "Scaleway: scaleway_vpc_public_gateway_ip"
subcategory: "VPC Gateway"
description: |-
  Lists Public Gateway IP addresses across zones and projects.
---

# Resource: scaleway_vpc_public_gateway_ip

For more information, see [the main documentation](https://www.scaleway.com/en/docs/network/vpc/how-to/create-a-public-gateway/).

## Example Usage

```terraform
# List Public Gateway IPs across all projects in a zone
list "scaleway_vpc_public_gateway_ip" "by_project" {
  provider = scaleway

  config {
    zones       = ["fr-par-1"]
    project_ids = ["*"]
  }
}
```

```terraform
# List only free (unattached) Public Gateway IPs
list "scaleway_vpc_public_gateway_ip" "free" {
  provider = scaleway

  config {
    zones   = ["*"]
    is_free = true
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `tags` - (Optional) Tags to filter for.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.
- `reverse` - (Optional) Filter for IPs whose reverse DNS contains this substring.
- `is_free` - (Optional) Filter based on whether the IP is attached to a gateway or not.

## Attributes Reference

Each listed item exposes the same attributes as the `scaleway_vpc_public_gateway_ip` managed resource.
