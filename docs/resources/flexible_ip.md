---
page_title: "Scaleway: scaleway_flexible_ip"
description: |-
  Manages Scaleway Flexible IPs.
---

# scaleway_flexible_ip

Creates and manages Scaleway flexible IPs.
For more information, see [the documentation](https://developers.scaleway.com/en/products/flexible-ip/api).

## Examples

### Basic

```hcl
resource "scaleway_flexible_ip" "main" {
    reverse = "my-reverse.com"
}
```

## Arguments Reference

The following arguments are supported:

- `description`: (Optional) A description of the flexible IP.
- `tags`: (Optional) A list of tags to apply to the flexible IP.
- `reverse` - (Optional) The reverse domain associated with this flexible IP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Flexible IP
- `ip_address` -  The IP address of the Flexible IP
- `zone` - The zone of the Flexible IP
- `organization_id` - The organization of the Flexible IP
- `project_id` - The project of the Flexible IP

## Import

Flexible IPs can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_flexible_ip.main fr-par-1/11111111-1111-1111-1111-111111111111
```
