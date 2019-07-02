---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_server"
sidebar_current: "docs-scaleway-resource-compute-instance-server"
description: |-
  Manages Scaleway Compute Instance servers.
---

# scaleway_compute_instance_server

Creates and manages Scaleway Compute Instance servers. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#servers-8bf7d7).

## Example Usage

```hcl
resource "scaleway_compute_instance_server" "web" {
  type = "DEV1-S"
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
}
```

## Arguments Reference

The following arguments are supported:

- `type` - The server commercial type. You can use [this endpoint](https://api.scaleway.com/instance/v1/zones/fr-par-1/products/servers) to find all the available commercial types. <!-- TODO: Add a link to an adapted tool -->
- `image_id` - The base image of the server. You can use [this endpoint](https://api-marketplace.scaleway.com/images?page=1&per_page=100) to find the right local image `ID` for a given image `name` and a given `commercial_type`. <!-- TODO: Add a link to an adapted tool -->
- `name` - (Optional) The name of the server.
- `tags` - (Optional) The tags associated with the server.
- `security_group_id` - The security group the server is attached to. <!-- TODO: Add a link to instance_compute_security_group -->
- `root_volume` - (Optional) Root volume attached to the server on creation.
  - `size_in_gb` - (Optional) Size of the root volume in gigabytes.
  - `delete_on_termination` - (Defaults to `false`) Force deletion of the root volume on instance termination.
- `additional_volume_ids` - (Optional) The additional volumes attached to the server.
- `enable_ipv6` - (Default to `false`) Determines if IPv6 is enabled for the server.
- `state` - (Default to `started`) The state of the server. Possible values are: `started`, `stopped` or `standby`.
- `cloud_init` - (Optional) The cloud init script associated with this server.
- `user_data` - (Optional) The user data associated with the server.
  - `key` - The user data key. The `cloud-init` key is reserved, please use `cloud_init` attribute instead.
  - `value` - The user data content. It could be a string or a file content using [file](https://www.terraform.io/docs/configuration/functions/file.html) or [filebase64](https://www.terraform.io/docs/configuration/functions/filebase64.html) for example.
- `zone` - (Optional) The [zone](https://developers.scaleway.com/en/quickstart/#zone-definition) in which the server should be created. If it is not provided, the provider `zone` is used.
- `project_id` - (Optional) The ID of the project the server is associated with. If it is not provided, the provider `project_id` is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the server.
- `root_volume`
  - `volume_id` - The volume ID of the root volume or the server.
- `private_ip` - The Scaleway internal IP address of the server.
- `public_ip` - The public IPv4 address of the server.

## Import

Instance servers can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_compute_instance_server.web fr-par-1/11111111-1111-1111-1111-111111111111
```
