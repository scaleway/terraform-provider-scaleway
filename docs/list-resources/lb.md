---
page_title: "Scaleway: scaleway_lb"
subcategory: "Load Balancers"
description: |-
  Lists Scaleway Load Balancers across zones and projects.
---

# Resource: scaleway_lb

Lists Scaleway Load Balancers across zones and projects.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/load-balancer/concepts/).

## Example Usage

```terraform
# List Load Balancers across all zones and all projects
list "scaleway_lb" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List Load Balancers across all zones filtered by name
list "scaleway_lb" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-lb"
  }
}
```

```terraform
# List Load Balancers filtered by tag
list "scaleway_lb" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["production"]
  }
}
```

```terraform
# List Load Balancers in a specific zone
list "scaleway_lb" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the Load Balancer to filter for.
- `tags` - (Optional) Tags to filter for.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Load Balancer:

- `id` - The ID of the Load Balancer.
- `name` - The name of the Load Balancer.
- `description` - The description of the Load Balancer.
- `type` - The type of the Load Balancer (e.g. `LB-S`).
- `tags` - The tags associated with the Load Balancer.
- `zone` - The zone of the Load Balancer.
- `region` - The region of the Load Balancer.
- `organization_id` - The ID of the organization the Load Balancer is associated with.
- `project_id` - The ID of the project the Load Balancer is associated with.
- `ip_id` - The ID of the primary IP attached to the Load Balancer.
- `ip_ids` - The IDs of all IPs attached to the Load Balancer.
- `ip_address` - The IPv4 address of the Load Balancer.
- `ipv6_address` - The IPv6 address of the Load Balancer.
- `ssl_compatibility_level` - The SSL compatibility level of the Load Balancer.
- `private_network` - The Private Networks attached to the Load Balancer.
- `private_ips` - The private IPs of the Load Balancer.
