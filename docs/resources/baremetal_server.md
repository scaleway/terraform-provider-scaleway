---
page_title: "Scaleway: scaleway_baremetal_server"
description: |-
  Manages Scaleway Compute Baremetal servers.
---

# scaleway_baremetal_server

Creates and manages Scaleway Compute Baremetal servers. For more information, see [the documentation](https://developers.scaleway.com/en/products/baremetal/api).

## Examples

### Basic

```hcl
data "scaleway_account_ssh_key" "main" {
  name = "main"
}

resource "scaleway_baremetal_server" "base" {
  zone		  = "fr-par-2"
  offer       = "GP-BM1-S"
  os          = "d17d6872-0412-45d9-a198-af82c34d3c5c"
  ssh_key_ids = [data.scaleway_account_ssh_key.main.id]
}
```

## Arguments Reference

The following arguments are supported:

- `offer` - (Required) The offer name or UUID of the baremetal server.
  Use [this endpoint](https://developers.scaleway.com/en/products/baremetal/api/#get-334154) to find the right offer.

~> **Important:** Updates to `offer` will recreate the server.

- `os` - (Required) The UUID of the os to install on the server.
  Use [this endpoint](https://developers.scaleway.com/en/products/baremetal/api/#get-87598a) to find the right OS ID.
  ~> **Important:** Updates to `os` will reinstall the server.
- `ssh_key_ids` - (Required) List of SSH keys allowed to connect to the server.
- `user` - (Optional) User used for the installation.
- `password` - (Optional) Password used for the installation. May be required depending on used os.
- `service_user` - (Optional) User used for the service to install.
- `service_password` - (Optional) Password used for the service to install. May be required depending on used os.
- `reinstall_on_config_changes` - (Optional) If True, this boolean allows to reinstall the server on install config changes.
  ~> **Important:** Updates to `ssh_key_ids`, `user`, `password`, `service_user` or `service_password` will not take effect on the server, it requires to reinstall it. To do so please set 'reinstall_on_config_changes' argument to true.
- `name` - (Optional) The name of the server.
- `hostname` - (Optional) The hostname of the server.
- `description` - (Optional) A description for the server.
- `tags` - (Optional) The tags associated with the server.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the server is associated with.


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the server.
- `offer_id` - The ID of the offer.
- `os_id` - The ID of the os.
- `ips` - (List of) The IPs of the server.
    - `id` - The ID of the IP.
    - `address` - The address of the IP.
    - `reverse` - The reverse of the IP.
    - `type` - The type of the IP.
- `domain` - The domain of the server.
- `organization_id` - The organization ID the server is associated with.

## Import

Baremetal servers can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_baremetal_server.web fr-par-2/11111111-1111-1111-1111-111111111111
```
