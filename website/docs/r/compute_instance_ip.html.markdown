---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_ip"
sidebar_current: "docs-scaleway-resource-compute-instance-ip"
description: |-
  Manages Scaleway Compute Instance IPs.
---

# scaleway_compute_instance_ip

Creates and manages Scaleway Compute Instance IPs. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#ips-268151).

## Example Usage

```hcl
resource "scaleway_compute_instance_ip" "server_ip" {}
```

## Arguments Reference

The following arguments are supported:

- `address` - (Computed) The IP address.
- `reverse` - (Optional) The reverse DNS for this IP.
- `project_id` - (Optional) The ID of the project you want to attach this resource to. If it is not provided, the provider `project_id` is used.
- `zone` - (Optional) The [zone](https://developers.scaleway.com/en/quickstart/#zone-definition) you want to attach this resource to. If it is not provided, the provider `zone` is used.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the IP.
- `address` - The IP address.
- `reverse` - The reverse DNS for this IP.
- `server_id` - The ID of the server this resource is attached to.
- `project_id` - The ID of the project you want to attach this resource to.
- `zone` - The [zone](https://developers.scaleway.com/en/quickstart/#zone-definition) your resource is attached to.


## Import

Instances can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_compute_instance_ip.server_ip fr-par-1/11111111-1111-1111-1111-111111111111
```
