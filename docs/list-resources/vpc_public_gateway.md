---
page_title: "Scaleway: scaleway_vpc_public_gateway"
subcategory: "VPC Gateway"
description: |-
  Lists Scaleway VPC Public Gateways across zones and projects.
---

# Resource: scaleway_vpc_public_gateway

Lists Scaleway VPC Public Gateways across zones and projects.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/network/vpc/concepts/#public-gateways).

## Example Usage

```terraform
# List Public Gateways across all zones and all projects
list "scaleway_vpc_public_gateway" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List Public Gateways filtered by tag
list "scaleway_vpc_public_gateway" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["prod"]
  }
}
```

```terraform
# List Public Gateways filtered by gateway type
list "scaleway_vpc_public_gateway" "by_type" {
  provider = scaleway

  config {
    zones = ["*"]
    types = ["VPC-GW-S"]
  }
}
```

```terraform
# List Public Gateways in a specific zone
list "scaleway_vpc_public_gateway" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the Public Gateway to filter for.
- `tags` - (Optional) Tags to filter for.
- `types` - (Optional) Filter for gateways of these types (e.g. `VPC-GW-S`, `VPC-GW-M`).
- `private_network_ids` - (Optional) Filter for gateways attached to these Private Networks.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Public Gateway:

- `id` - The ID of the Public Gateway.
- `name` - The name of the Public Gateway.
- `type` - The type of the Public Gateway.
- `status` - The current status of the Public Gateway.
- `zone` - The zone of the Public Gateway.
- `created_at` - The date and time of the creation of the Public Gateway.
- `updated_at` - The date and time of the last update of the Public Gateway.
- `organization_id` - The ID of the organization the Public Gateway is associated with.
- `project_id` - The ID of the project the Public Gateway is associated with.
- `tags` - The tags associated with the Public Gateway.
- `ip_id` - The ID of the IP address associated with the Public Gateway.
- `bandwidth` - The bandwidth of the Public Gateway (in Mbps).
- `bastion_enabled` - Whether bastion is enabled on the Public Gateway.
- `bastion_port` - The port of the bastion.
- `enable_smtp` - Whether SMTP is enabled on the Public Gateway.
