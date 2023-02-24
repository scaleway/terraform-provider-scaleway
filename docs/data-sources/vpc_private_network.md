---
page_title: "Scaleway: scaleway_vpc_private_network"
description: |-
  Get information about Scaleway VPC Private Networks.
---

# scaleway_vpc_private_network

Gets information about a private network.

## Example Usage

N/A, the usage will be meaningful in the next releases of VPC.

## Argument Reference

* `name` - (Required) Exact name of the private network.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the private network.

~> **Important:** Private networks' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

