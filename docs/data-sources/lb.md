---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb"
---

# scaleway_lb

Gets information about a Load Balancer.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/concepts/#load-balancers) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-load-balancer-list-load-balancers).

## Example Usage

```hcl
# Get info by name
data "scaleway_lb" "by_name" {
  name = "foobar"
}

# Get info by ID
data "scaleway_lb" "by_id" {
  lb_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The Load Balancer name.

- `ip_id` - (Optional) The Load Balancer IP ID.

- `project_id` - (Optional) The ID of the Project the Load Balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Load Balancer.

~> **Important:** Load Balancer IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `ip_address` - The Load Balancer public IP address.

- `type` - The Load Balancer type.

- `tags` - The tags associated with the Load Balancer.

- `zone` -  (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Load Balancer exists.
