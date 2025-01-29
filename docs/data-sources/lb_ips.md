---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_ips"
---

# scaleway_lb_ips

Gets information about multiple Load Balancer IP addresses.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/how-to/create-manage-flex-ips/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-ip-addresses-list-ip-addresses).

## Example Usage

```hcl
# Find multiple IPs that share the same CIDR block
data "scaleway_lb_ips" "my_key" {
  ip_cidr_range = "0.0.0.0/0"
}
# Find IPs by CIDR block and zone
data "scaleway_lb_ips" "my_key" {
  ip_cidr_range = "0.0.0.0/0"
  zone       = "fr-par-2"
}

# Find IPs that share the same tags and type
data "scaleway_lb_ips" "ips_by_tags_and_type" {
  tags    = [ "a tag" ]
  ip_type = "ipv4"
}
```

## Argument Reference

- `ip_cidr_range` - (Optional) The IP CIDR range to filter for. IPs within a matching CIDR block are listed.

- `tags` - (Optional)  List of tags used as filter. IPs with these exact tags are listed.

- `ip_type` - (Optional) The IP type used as a filter.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IPs exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `ips` - List of retrieved IPs
    - `id` - The ID of the associated IP.
    - `lb_id` - The ID of the associated Load BalancerD, if any
    - `ip_address` - The IP address
    - `zone` - The [zone](../guides/regions_and_zones.md#zones) of the Load Balancer.
    - `reverse` - The reverse domain associated with this IP.
    - `organization_id` - The ID of the Organization the Load Balancer is associated with.
    - `project_id` - The ID of the Project the Load Balancer is associated with.