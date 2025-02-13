---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_ip"
---

# scaleway_lb_ip

Gets information about a Load Balancer IP address.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/how-to/create-manage-flex-ips/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-ip-addresses-list-ip-addresses).


## Example Usage

```hcl
# Get info by IP address
data "scaleway_lb_ip" "my_ip" {
  ip_address = "0.0.0.0"
}

# Get info by IP ID
data "scaleway_lb_ip" "my_ip" {
  ip_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

The following arguments are supported:

- `ip_address` - (Optional) The IP address.
  Only one of `ip_address` and `ip_id` should be specified.

- `ip_id` - (Optional) The IP ID.
  Only one of `ip_address` and `ip_id` should be specified.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP was reserved.

- `project_id` - (Optional) The ID of the Project the Load Balancer IP is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IP.

~> **Important:** Load Balancers IP IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `reverse` - The reverse domain associated with this IP.

- `lb_id` - The ID of the associated Load Balancer, if any

- `tags` - The tags associated with this IP.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the Organization the Load Balancer IP is associated with.
