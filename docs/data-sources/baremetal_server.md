---
page_title: "Scaleway: scaleway_baremetal_server"
description: |-
  Gets information about a baremetal server.
---

# scaleway_baremetal_server

Gets information about a baremetal server.
For more information, see [the documentation](https://developers.scaleway.com/en/products/baremetal/api).

## Example Usage

```hcl
# Get info by server name
data "scaleway_baremetal_server" "by_name" {
  name = "foobar"
  zone = "fr-par-2"
}

# Get info by server id
data "scaleway_baremetal_server" "by_id" {
  server_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The server name. Only one of `name` and `server_id` should be specified.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server exists.
