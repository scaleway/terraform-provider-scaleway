---
page_title: "Scaleway: scaleway_instance_servers"
description: |-
Gets information about multiple Instance Servers.
---

# scaleway_instance_servers

Gets information about multiple instance servers.

## Examples

### Basic

```hcl
# Find servers by tag
data "scaleway_instance_servers" "my_key" {
  tags  = ["tag"]
}

# Find servers by name and zone
data "scaleway_instance_servers" "my_key" {
  name = "myserver"
  zone = "fr-par-2"
}
```

### With single datasource

```hcl
# Find servers by tag
data "scaleway_instance_servers" "servers_by_tag" {
  tags  = ["tag"]
}

data "scaleway_instance_servers" "map_of_servers" {
  for_each = {for server in data.scaleway_instance_servers.servers_by_tag.servers: server.id => server}
  server_id = each.value.id
}
```

## Argument Reference

- `name` - (Optional) The server name used as filter.

- `tags` - (Optional) List of tags used as filter.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which servers exist.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The zone of the servers

- `servers` - List of found servers
  - `id` - The ID of the server.
  - `public_ip` - The public IPv4 address of the server.
