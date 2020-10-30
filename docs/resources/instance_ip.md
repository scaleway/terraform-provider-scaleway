---
page_title: "Scaleway: scaleway_instance_ip"
description: |-
  Manages Scaleway Compute Instance IPs.
---

# scaleway_instance_ip

Creates and manages Scaleway Compute Instance IPs. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#ips-268151).

## Example Usage

```hcl
resource "scaleway_instance_ip" "server_ip" {}
```

## Arguments Reference

The following arguments are supported:

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.
- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the IP is associated with.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the IP is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IP.
- `address` - The IP address.
- `reverse` - The reverse dns attached to this IP

## Import

IPs can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_ip.server_ip fr-par-1/11111111-1111-1111-1111-111111111111
```
