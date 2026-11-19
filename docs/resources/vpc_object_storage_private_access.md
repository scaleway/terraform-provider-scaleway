---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_object_storage_private_access"
---

# Resource: scaleway_vpc_object_storage_private_access

Object Storage private access lets resources attached to selected Private Networks reach Scaleway Object Storage over the private network, without using the public Internet.



## Example Usage

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_vpc_object_storage_private_access" "main" {
  vpc_id = scaleway_vpc.vpc01.id

  private_network_ids = [
    scaleway_vpc_private_network.pn01.id,
  ]
}
```

```terraform
resource "scaleway_vpc" "vpc01" {
  name   = "my-vpc"
  region = "nl-ams"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
  region = "nl-ams"
}

resource "scaleway_vpc_object_storage_private_access" "main" {
  vpc_id = scaleway_vpc.vpc01.id
  region = "nl-ams"

  private_network_ids = [
    scaleway_vpc_private_network.pn01.id,
  ]
}
```



## Argument Reference

- `vpc_id` - (Required) The ID of the VPC for which Object Storage private access is enabled. Changing this forces a new resource.
- `private_network_ids` - (Required) IDs of the Private Networks that can reach Object Storage over the private network. At least one Private Network is required.
- `region` - (Optional, Computed) The [region](../guides/regions_and_zones.md#regions) of the VPC. Defaults to the region configured on the provider. Changing this forces a new resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Object Storage private access, in the `{region}/{vpc_id}` format.

## Import

Object Storage private access can be imported using the `{region}/{vpc_id}` identifier:

```bash
terraform import scaleway_vpc_object_storage_private_access.main fr-par/11111111-1111-1111-1111-111111111111
```
