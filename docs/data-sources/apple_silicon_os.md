---
subcategory: "Apple Silicon"
page_title: "Scaleway: scaleway_apple_silicon_os"
---

# scaleway_apple_silicon_os

Gets information about a Apple Silicon operating system.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/apple-silicon/#path-os-list-all-operating-systems-os).

You can also use the [scaleway-cli](https://github.com/scaleway/scaleway-cli) with `scw apple-silicon os list` to list all available operating systems.

## Example Usage

```hcl
# Get info by os name and version
data "scaleway_apple_silicon_os" "by_name" {
  name = "devos-sequoia-15.6"
}

# Get info by os id
data "scaleway_apple_silicon_os" "by_id" {
  os_id = "cafecafe-5018-4dcd-bd08-35f031b0ac3e"
}
```

## Argument Reference

- `name` - (Optional) The os name. Only one of `name` and `os_id` should be specified.
- `version` - (Optional) The os version.
- `os_id` - (Optional) The operating system id. Only one of `name` and `os_id` should be specified.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the os exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The resource's ID

~> **Important:** Apple Silicon operating systems' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
