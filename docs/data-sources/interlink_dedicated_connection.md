---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_dedicated_connection"
---

# scaleway_interlink_dedicated_connection

Gets information about an Interlink Dedicated Connection.

A dedicated connection is a physical connection owned by the user at a PoP, used to create self-hosted links between your infrastructure and Scaleway.



## Example Usage

```terraform
# Retrieve a dedicated connection by its ID
data "scaleway_interlink_dedicated_connection" "by_id" {
  connection_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve a dedicated connection by name
data "scaleway_interlink_dedicated_connection" "by_name" {
  name = "my-dedicated-connection"
}
```




## Argument Reference

- `connection_id` - (Optional) The ID of the dedicated connection. Can be a plain UUID or a regional ID. Conflicts with `name`.
- `name` - (Optional) The name of the dedicated connection to filter for. Conflicts with `connection_id`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the dedicated connection exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the dedicated connection.
- `status` - Status of the dedicated connection.
- `tags` - List of tags associated with the dedicated connection.
- `pop_id` - ID of the PoP where the dedicated connection is located.
- `bandwidth_mbps` - Bandwidth size of the dedicated connection in Mbps.
- `available_link_bandwidths` - Sizes of the links supported on this dedicated connection.
- `demarcation_info` - Demarcation details required by the data center to set up the Cross Connect.
- `vlan_range` - VLAN range for self-hosted links. Contains `start` and `end`.
- `project_id` - The ID of the project the dedicated connection belongs to.
- `organization_id` - The ID of the organization the dedicated connection belongs to.
- `created_at` - Creation date of the dedicated connection (RFC 3339 format).
- `updated_at` - Last modification date of the dedicated connection (RFC 3339 format).
