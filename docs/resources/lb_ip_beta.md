---
layout: "scaleway"
page_title: "Scaleway: scaleway_lb_ip_beta"
description: |-
  Manages Scaleway Load-Balancers IPs.
---

# scaleway_lb_ip_beta

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancers IPs. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples
    
### Basic

```hcl
resource "scaleway_lb_ip_beta" "ip" {
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

IPs can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_ip_beta.ip01 fr-par/11111111-1111-1111-1111-111111111111
```
