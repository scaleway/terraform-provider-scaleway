---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# scaleway_vpc_private_network

Creates and manages Scaleway VPC Private Networks.
For more information, see [the documentation](https://developers.scaleway.com/en/products/vpc/api/#private-networks-ac2df4).

## Example

```hcl
resource "scaleway_vpc_private_network" "pn_priv" {
    name = "subnet_demo"
    tags = ["demo", "terraform"]
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the private network. If not provided it will be randomly generated.
- `tags` - (Optional) The tags associated with the private network.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the private network should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the private network is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the private network.

~> **Important:** Private networks' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `organization_id` - The organization ID the private network is associated with.

## Import

Private networks can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_vpc_private_network.vpc_demo fr-par-1/11111111-1111-1111-1111-111111111111
```
