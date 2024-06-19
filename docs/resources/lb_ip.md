---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_ip"
---

# Resource: scaleway_lb_ip

Creates and manages Scaleway Load Balancer IP addresses.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/network/load-balancer/how-to/create-manage-flex-ips/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-ip-addresses-list-ip-addresses).

## Example Usage

### Basic

```terraform
resource "scaleway_lb_ip" "ip" {
    reverse = "my-reverse.com"
}
```

### With IPv6

```terraform
resource "scaleway_lb_ip" "ipv6" {
    is_ipv6 = true
}
```

## Argument Reference

The following arguments are supported:

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the IP is associated with.
- `reverse` - (Optional) The reverse domain associated with this IP.
- `is_ipv6` - (Optional) If true, creates a flexible IP with an IPv6 address.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the IP address

~> **Important:** Load-Balancer IP IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `lb_id` - The associated Load Balancer ID if any
- `ip_address` -  The IP address

## Import

IPs can be imported using `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_ip.ip01 fr-par-1/11111111-1111-1111-1111-111111111111
```
