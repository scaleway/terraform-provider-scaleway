---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_ingress_rule"
---

# scaleway_vpc_ingress_rule (Data Source)

Gets information about a VPC ingress rule.

An ingress routing rule routes incoming traffic from a peered VPC to a specific private IP address within a destination VPC's Private Network.



## Example Usage

```terraform
# Retrieve a VPC ingress rule by filters
data "scaleway_vpc_ingress_rule" "by_pn" {
  nexthop_private_network_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve a VPC ingress rule by its ID
data "scaleway_vpc_ingress_rule" "by_id" {
  ingress_rule_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```




## Argument Reference

- `ingress_rule_id` - (Optional) The ID of the VPC ingress rule. Conflicts with all filter arguments below.

The following arguments can be used to look up a VPC ingress rule via the list API. They all conflict with `ingress_rule_id`:

- `vpc_id` - (Optional) The VPC ID to filter for.
- `nexthop_resource_ip` - (Optional) The nexthop resource IP to filter for.
- `nexthop_private_network_id` - (Optional) The nexthop private network ID to filter for.
- `is_ipv6` - (Optional) Only ingress rules with the matching IP version will be returned.
- `tags` - (Optional) List of tags to filter for.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the ingress rule exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_vpc_ingress_rule` [resource](../resources/vpc_ingress_rule.md).
