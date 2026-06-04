---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_acl"
---

# scaleway_vpc_acl

Gets information about a VPC ACL (Access Control List).


## Example Usage

```terraform
# Get the IPv4 ACL for a VPC
data "scaleway_vpc_acl" "my_acl" {
  vpc_id  = scaleway_vpc.my_vpc.id
  is_ipv6 = false
}
```

```terraform
# Get the IPv6 ACL for a VPC
data "scaleway_vpc_acl" "my_acl_v6" {
  vpc_id  = scaleway_vpc.my_vpc.id
  is_ipv6 = true
}
```



## Argument Reference

- `vpc_id` - (Required) The VPC ID to look up the ACL for.

- `is_ipv6` - (Optional, defaults to `false`) Whether to get the IPv6 ACL.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the ACL exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_vpc_acl` [resource](../resources/vpc_acl.md).
