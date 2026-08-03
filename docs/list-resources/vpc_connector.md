---
page_title: "Scaleway: scaleway_vpc_connector"
subcategory: "VPC"
description: |-
  Lists Scaleway VPC Connectors across regions and projects.
---

# Resource: scaleway_vpc_connector

For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).


## Example Usage

```terraform
# List VPC connectors across all regions and all projects
list "scaleway_vpc_connector" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```
```terraform
# List VPC connectors filtered by name (matches connectors whose name contains "prod")
list "scaleway_vpc_connector" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "prod"
  }
}
```
```terraform
# List VPC connectors attached to a specific source VPC
list "scaleway_vpc_connector" "by_vpc" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    vpc_id  = "11111111-1111-1111-1111-111111111111"
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the VPC connector to filter for.
- `tags` - (Optional) Tags to filter for.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.
- `vpc_id` - (Optional) Filter for connectors attached to this source VPC (regional ID or plain UUID).
- `target_vpc_id` - (Optional) Filter for connectors attached to this target VPC (regional ID or plain UUID).

## Attributes Reference

In addition to the arguments above, each listed VPC connector exports the same attributes as the `scaleway_vpc_connector` managed resource:

- `id` - The ID of the VPC connector.
- `name` - The name of the VPC connector.
- `vpc_id` - The ID of the source VPC.
- `target_vpc_id` - The ID of the target VPC.
- `status` - The VPC connector status.
- `region` - The region of the VPC connector.
- `organization_id` - The ID of the organization the VPC connector is associated with.
- `project_id` - The ID of the project the VPC connector is associated with.
- `tags` - The tags associated with the VPC connector.
- `created_at` - The date and time of the creation of the VPC connector.
- `updated_at` - The date and time of the last update of the VPC connector.
