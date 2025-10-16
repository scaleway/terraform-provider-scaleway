---
subcategory: "Instances"
page_title: "Scaleway: scaleway_instance_server_type"
---

# scaleway_instance_server_type

Gets information about a server type.

## Example Usage

```hcl
data "scaleway_instance_server_type" "pro2-s" {
  name = "PRO2-S"
  zone = "nl-ams-1"
}
```

## Argument Reference

To select the server type which information should be fetched, the following arguments can be used:

- `name` - (Required) The name of the server type.
  Only one of `name` and `snapshot_id` should be specified.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) of the server type (to check the availability of the server type for example).

## Attributes Reference

The following attributes will be available:

- `arch` - The architecture of the server type.

- `cpu` - The number of CPU cores of the server type.

- `ram` - The amount of RAM of the server type (in bytes).

- `gpu` - The number of GPUs of the server type.

- `volumes` - The specifications of volumes allowed for the server type.

    -> The `volumes` block contains:
    - `min_size_total` - The minimum total size in bytes of volumes allowed on the server type.
    - `max_size_total` - The maximum total size in bytes of volumes allowed on the server type.
    - `min_size_per_local_volume` - The minimum size in bytes per local volume allowed on the server type.
    - `max_size_per_local_volume` - The maximum size in bytes per local volume allowed on the server type.
    - `scratch_storage_max_size` - The maximum size in bytes of the scratch volume allowed on the server type.
    - `block_storage` - Whether block storage is allowed on the server type.

- `capabilities` - The specific capabilities of the server type.

    -> The `capabilities` block contains:
    - `boot_types` - The boot types allowed for the server type.
    - `max_file_systems` - The maximum number of file systems that can be attached on the server type.

- `network` - The network specifications of the server type.

    -> The `network` block contains:
    - `internal_bandwidth` - The internal bandwidth of the server type (in bytes/second).
    - `public_bandwidth` - The public bandwidth of the server type (in bytes/second).
    - `block_bandwidth` - The block bandwidth of the server type (in bytes/second).

- `hourly_price` - The hourly price of the server type (in euros).

- `monthly_price` - The monthly price of the server type (in euros).

- `end_of_service` - Whether the server type will soon reach End Of Service.

- `availability` - Whether the server type is available in the zone.
