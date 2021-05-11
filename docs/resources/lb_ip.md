---
page_title: "Scaleway: scaleway_lb_ip"
description: |-
  Manages Scaleway Load-Balancers IPs.
---

# scaleway_lb_ip

Creates and manages Scaleway Load-Balancers IPs.
For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/zoned_api).

## Examples

### Basic

```hcl
resource "scaleway_lb_ip" "ip" {
    reverse = "my-reverse.com"
}
```

## Arguments Reference

The following arguments are supported:

- `reverse` - (Optional) The reverse domain associated with this IP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the IP
- `lb_id` - The associated load-balance ID if any
- `ip_address` -  The IP Address


## Import

IPs can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_ip.ip01 fr-par-1/11111111-1111-1111-1111-111111111111
```
