---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_ingress_rule"
---

# Resource: scaleway_vpc_ingress_rule

Creates and manages Scaleway VPC Ingress Rules.

An ingress routing rule routes incoming traffic from a peered VPC to a specific private IP address within a destination VPC's Private Network.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).



## Example Usage

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_vpc_ingress_rule" "main" {
  vpc_id                     = scaleway_vpc.vpc01.id
  source                     = "10.0.0.0/24"
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
  nexthop_resource_ip        = "10.0.0.10"
  description                = "Allow ingress traffic from 10.0.0.0/24"
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

resource "scaleway_vpc_ingress_rule" "main" {
  vpc_id                     = scaleway_vpc.vpc01.id
  source                     = "10.0.0.0/24"
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
  nexthop_resource_ip        = "10.0.0.10"
  region                     = "nl-ams"
}
```

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_vpc_ingress_rule" "main" {
  vpc_id                     = scaleway_vpc.vpc01.id
  source                     = "10.0.0.0/24"
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
  nexthop_resource_ip        = "10.0.0.10"
  description                = "Allow ingress traffic from 10.0.0.0/24"
  tags                       = ["production", "ingress"]
}
```




## Argument Reference

The following arguments are supported:

- `vpc_id` - (Required) The ID of the VPC in which to create the ingress rule.
- `source` - (Required) Source IP range (in CIDR notation) to which the ingress rule applies.
- `nexthop_resource_ip` - (Required) IP of the nexthop resource that should handle traffic matched by this rule.
- `nexthop_private_network_id` - (Required) The ID of the private network used as nexthop for traffic matched by this rule.
- `description` - (Optional) The description of the ingress rule.
- `tags` - (Optional) The tags to associate with the ingress rule.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) of the ingress rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the VPC ingress rule.
- `is_ipv6` - Whether the ingress rule is for IPv6 traffic (derived from `source`).
- `created_at` - The date and time of the creation of the ingress rule (RFC 3339 format).
- `updated_at` - The date and time of the last update of the ingress rule (RFC 3339 format).
- `srn` - The Scaleway Resource Name (SRN) of the ingress rule.

~> **Important:** VPC ingress rules' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

VPC ingress rules can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_vpc_ingress_rule.main fr-par/11111111-1111-1111-1111-111111111111
```
