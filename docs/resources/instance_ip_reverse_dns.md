---
page_title: "Scaleway: scaleway_instance_ip_reverse_dns"
description: |-
  Manages Scaleway Compute Instance IPs' reverse DNS.
---

# scaleway_instance_ip_reverse_dns

Manages Scaleway Compute Instance IPs Reverse DNS.

## Example Usage

```hcl
resource "scaleway_instance_ip" "server_ip" {}

resource "scaleway_instance_ip_reverse_dns" "reverse" {
  ip_id   = scaleway_instance_ip.server_ip.id
  reverse = "www.scaleway.com"
}
```

## Arguments Reference

The following arguments are supported:

- `ip_id` - (Required) The IP ID
- `reverse` - (Required) The reverse DNS for this IP.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IP.

## Import

IPs reverse DNS can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_ip_reverse_dns.reverse fr-par-1/11111111-1111-1111-1111-111111111111
```
