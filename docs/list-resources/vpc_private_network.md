---
page_title: "Scaleway: scaleway_vpc_private_network"
subcategory: "VPC"
description: |-
  Lists Scaleway VPC Private Networks across regions and projects.
---

# Resource: scaleway_vpc_private_network



For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/#private-networks).


## Example Usage

```terraform
# List Private Networks across all regions and all projects
list "scaleway_vpc_private_network" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List Private Networks across all regions filtered by name prefix
list "scaleway_vpc_private_network" "by_name" {
  provider = scaleway

  config {
    regions     = ["*"]
    name        = "my-network"
  }
}
```

```terraform
# List Private Networks in a specific region (fr-par) for a specific project
list "scaleway_vpc_private_network" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List Private Networks filtered by a specific tag
list "scaleway_vpc_private_network" "by_tag" {
  provider = scaleway

  config {
    regions     = ["*"]
    tags        = ["production"]
  }
}
```

```terraform
# List all Private Networks belonging to a specific VPC
list "scaleway_vpc_private_network" "by_vpc" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    vpc_id  = "11111111-1111-1111-1111-111111111111"
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the Private Network to filter for.
- `tags` - (Optional) Tags to filter for.
- `vpc_id` - (Optional) VPC ID to filter for. Only Private Networks belonging to this VPC will be returned.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Private Network:

- `id` - The ID of the Private Network.
- `name` - The name of the Private Network.
- `created_at` - The date and time of the creation of the Private Network.
- `updated_at` - The date and time of the last update of the Private Network.
- `organization_id` - The ID of the organization the Private Network is associated with.
- `project_id` - The ID of the project the Private Network is associated with.
- `region` - The region of the Private Network.
- `vpc_id` - The ID of the VPC the Private Network belongs to.
- `tags` - The tags associated with the Private Network.
- `ipv4_subnet` - The IPv4 subnet associated with the Private Network.
- `ipv6_subnets` - The IPv6 subnets associated with the Private Network.
- `is_regional` - Whether the Private Network is regional.
- `enable_default_route_propagation` - Whether default route propagation is enabled.
