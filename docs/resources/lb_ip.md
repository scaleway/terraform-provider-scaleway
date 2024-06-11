---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_ip"
---

# Resource: scaleway_lb_ip

Creates and manages Scaleway Load-Balancers IPs.
For more information, see [the documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-ip-addresses).

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
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the IP is associated with.
- `reverse` - (Optional) The reverse domain associated with this IP.
- `is_ipv6` - (Optional) If true, creates a Flexible IP with an IPv6 address.
- `tags` - (Optional) The tags associated with this IP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the IP

~> **Important:** Load-Balancers IPs' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `lb_id` - The associated load-balance ID if any
- `ip_address` -  The IP Address

## Import

IPs can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_ip.ip01 fr-par-1/11111111-1111-1111-1111-111111111111
```
