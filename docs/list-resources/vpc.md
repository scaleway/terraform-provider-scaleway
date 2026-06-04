---
page_title: "Scaleway: scaleway_vpc"
subcategory: "VPC"
description: |-
  Lists Scaleway Virtual Private Clouds (VPCs) across regions and projects.
---

# Resource: scaleway_vpc



For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).


## Example Usage

```terraform
# List VPCs across all regions and all projects
list "scaleway_vpc" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List VPCs across all regions filtered by name prefix (matches VPCs with names starting with "test-vpc")
list "scaleway_vpc" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "test-vpc"
  }
}
```

```terraform
# List VPCs in a specific region (fr-par) for a specific project
list "scaleway_vpc" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List VPCs in all regions for the default project filtered by a specific tag
list "scaleway_vpc" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["foobar"]
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the VPC to filter for.
- `tags` - (Optional) Tags to filter for.
- `is_default` - (Optional) Whether to filter for the default VPC only.
- `routing_enabled` - (Optional) Whether to filter for VPCs with routing enabled.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each VPC:

- `id` - The ID of the VPC.
- `name` - The name of the VPC.
- `created_at` - The date and time of the creation of the VPC.
- `updated_at` - The date and time of the last update of the VPC.
- `organization_id` - The ID of the organization the VPC is associated with.
- `project_id` - The ID of the project the VPC is associated with.
- `region` - The region of the VPC.
- `is_default` - Whether the VPC is the default VPC.
- `tags` - The tags associated with the VPC.
- `enable_routing` - Whether routing is enabled for the VPC.
- `enable_custom_routes_propagation` - Whether custom routes propagation is enabled for the VPC.
