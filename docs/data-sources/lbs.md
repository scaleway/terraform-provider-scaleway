---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lbs"
---

# scaleway_lbs

Gets information about multiple Load Balancers.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/concepts/#load-balancers) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-load-balancer-list-load-balancers).

## Example Usage

```hcl
# Find LBs by name
data "scaleway_lbs" "my_key" {
  name = "foobar"
}

# Find LBs by name and zone
data "scaleway_lbs" "my_key" {
  name = "foobar"
  zone = "fr-par-2"
}

# Find LBs that share the same tags
data "scaleway_lbs" "lbs_by_tags" {
  tags = [ "a tag" ]
}
```

## Argument Reference

- `name` - (Optional) The Load Balancer name to filter for. Load Balancers with a matching name are listed.

- `tags` - (Optional)  List of tags to filter for. Load Balancers with these exact tags are listed.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Load Balancers exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `lbs` - List of retrieved Load Balancers
    - `id` - The ID of the Load Balancer.
    - `tags` - The tags associated with the Load Balancer.
    - `description` - The description of the Load Balancer.
    - `status` - The state of the Load Balancer Instance. Possible values are: `unknown`, `ready`, `pending`, `stopped`, `error`, `locked` and `migrating`.
    - `zone` - The [zone](../guides/regions_and_zones.md#zones) of the Load Balancer.
    - `name` - The name of the Load Balancer.
    - `type` - The offer type of the Load Balancer.
    - `instances` - List of underlying Instances.
    - `ips` - List of IPs attached to the Load Balancer.
    - `frontend_count` - Number of frontends the Load Balancer has.
    - `backend_count` - Number of backends the Load Balancer has.
    - `private_network_count` - Number of Private Networks attached to the Load balancer.
    - `route_count` - Number of routes the Load balancer has.
    - `subscriber` - The subscriber information.
    - `ssl_compatibility_level` - Determines the minimal SSL version which needs to be supported on the client side.
    - `created_at` - Date on which the Load Balancer was created.
    - `updated_at` - Date on which the Load Balancer was updated.
    - `organization_id` - The ID of the Organization the Load Balancer is associated with.
    - `project_id` - The ID of the Project the Load Balancer is associated with.
