---
subcategory: "IPAM"
page_title: "Scaleway: scaleway_ipam_ips"
---

# scaleway_ipam_ips

Gets information about multiple IP addresses managed by Scaleway's IP Address Management (IPAM) service.

For more information about IPAM, see the main [documentation](https://www.scaleway.com/en/docs/vpc/concepts/#ipam).

## Examples

### By tag

```terraform
data "scaleway_ipam_ips" "by_tag" {
  tags = ["tag"]
}
```

### By type and resource

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_redis_cluster" "redis01" {
  name         = "my_redis_cluster"
  version      = "7.0.5"
  node_type    = "RED1-XS"
  user_name    = "my_initial_user"
  password     = "thiZ_is_v&ry_s3cret"
  cluster_size = 3
  private_network {
    id = scaleway_vpc_private_network.pn01.id
  }
}

data "scaleway_ipam_ips" "by_type_and_resource" {
  type = "ipv4"
  resource {
    id   = scaleway_redis_cluster.redis01.id
    type = "redis_cluster"
  }
}
```

## Argument Reference

- `type` - (Optional) The type of IP to filter for (`ipv4` or `ipv6`).

- `private_network_id` - (Optional) The ID of the Private Network to filter for.

- `resource` - (Optional) Filter for a resource attached to the IP, using resource ID, type or name.
    - `id` - The ID of the attached resource.
    - `type` - The type of the attached resource. [Documentation](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/ipam/v1#pkg-constants) with type list.
    - `name` - The name of the attached resource.

- `mac_address` - (Optional) The linked MAC address to filter for.

- `tags` (Optional) The IP tags to filter for.

- `attached` - (Optional) Defines whether to filter only for IPs which are attached to a resource.

- `zonal` - (Optional) Only IPs that are zonal, and in this zone, will be returned.

- `region` - (Optional) The region to filter for.

- `project_id` - (Optional) The ID of the Project to filter for.

- `organization_id` - (Optional) The ID of the Organization to filter for.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The region of the IPs.
- `ips` - List of found IPs.
    - `id` - The ID of the IP.

    ~> **Important:** IPAM IP IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

    - `address` - The Scaleway internal IP address of the resource.
    - `resource` - The list of public IPs attached to the resource.
        - `id` - The ID of the resource.
        - `type` - The type of resource.
        - `mac_address` - The associated MAC address.
        - `name` - The name of the resource.
    - `tags` - The tags associated with the IP.
    - `created_at` - The date and time of the creation of the IP.
    - `updated_at` - The date and time of the last update of the IP.
    - `zone` - The [zone](../guides/regions_and_zones.md#zones) of the IP.
    - `region` - The [region](../guides/regions_and_zones.md#regions) of the IP.
    - `project_id` - The ID of the Project the resource is associated with.
  