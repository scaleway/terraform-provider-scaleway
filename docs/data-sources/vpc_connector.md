---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_connector"
---

# scaleway_vpc_connector (Data Source)

Gets information about a VPC connector.

A VPC connector enables network connectivity between two VPCs, allowing resources in separate VPCs to communicate with each other.

## Example Usage

```terraform
# Retrieve a VPC connector by its ID
data "scaleway_vpc_connector" "by_id" {
  connector_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve a VPC connector by name
data "scaleway_vpc_connector" "by_name" {
  name = "my-vpc-connector"
}
```



## Argument Reference

- `connector_id` - (Optional) The ID of the VPC connector. Conflicts with all filter arguments below.

The following arguments can be used to look up a VPC connector via the list API. They all conflict with `connector_id`:

- `name` - (Optional) The name to filter for.
- `vpc_id` - (Optional) The source VPC ID to filter for.
- `target_vpc_id` - (Optional) The target VPC ID to filter for.
- `tags` - (Optional) List of tags to filter for.
- `project_id` - (Optional) The ID of the Project to filter for.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the connector exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_vpc_connector` [resource](../resources/vpc_connector.md).
