---
page_title: "Scaleway: scaleway_baremetal_os"
description: |-
  Gets information about a baremetal operating system.
---

# scaleway_baremetal_os

Gets information about a baremetal operating system.
For more information, see [the documentation](https://developers.scaleway.com/en/products/baremetal/api).

## Example Usage

```hcl
# Get info by os name and version
data "scaleway_baremetal_os" "by_name" {
  name = "Ubuntu"
  version = "20.04"
}

# Get info by os id
data "scaleway_baremetal_os" "by_id" {
  os_id = "03b7f4ba-a6a1-4305-984e-b54fafbf1681"
}
```

## Argument Reference

- `name` - (Optional) The os name. Only one of `name` and `os_id` should be specified.
- `version` - (Optional) The os version.
- `os_id` - (Optional) The operating system id. Only one of `name` and `os_id` should be specified.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the os exists.
