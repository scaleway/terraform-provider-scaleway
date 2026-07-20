---
page_title: "Scaleway: scaleway_instance_server"
subcategory: "Instances"
description: |-
  Lists Scaleway Instance Servers across zones and projects.
---

# Resource: scaleway_instance_server

Lists Scaleway Instance Servers across zones and projects using query parameters.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/instances).

## Example Usage

```terraform
# List all servers of a specific project in all zones
list "scaleway_instance_server" "by_project" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List servers filtered by name and tag on a specific zone across all projects
list "scaleway_instance_server" "by_name_and_tag" {
  provider = scaleway

  config {
    zones       = ["fr-par-1"]
    project_ids = ["*"]
    name        = "the-exact-name"
    tags        = ["a-specific-tag"]
  }
}
```

```terraform
# List servers filtered by server-type on the default project across all zones
list "scaleway_instance_server" "by_name_and_tag" {
  provider = scaleway

  config {
    zones       = ["*"]
    server_type = "PRO2-L"
  }
}
```

```terraform
# List servers of a specific zone matching the filters (on the default project)
list "scaleway_instance_server" "by_project" {
  provider = scaleway

  config {
    zones = ["nl-ams-2"]
    security_group_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    placement_group_ids = [scaleway_instance_placement_group.pg.id]
    private_network_ids = [
      scaleway_vpc_private_network.pn1.id,
      scaleway_vpc_private_network.pn2.id,
    ]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `zones` - (Optional) Zones to filter by. Use `["*"]` to list servers from all zones.
- `project_ids` - (Optional) List of Project IDs to filter by. Use `["*"]` to list servers from all projects.
- `name` - (Optional) Name of the server to filter by. The name-matching is strict, and will not pick up servers which names only contain the `name` parameter.
- `tags` - (Optional) List of tags of the server to filter by.
- `server_type` - (Optional) Commercial type of the server to filter by. The name-matching is strict, and will not pick up servers which type only contain the `server-type` parameter.
- `security_group_ids` - (Optional) List of Security Group IDs to filter the servers by. Accepts either zonal IDs (e.g. `fr-par-1/uuid`) or plain UUIDs.
- `placement_group_ids` - (Optional) List of Placement Group IDs to filter the servers by. Accepts either zonal IDs (e.g. `fr-par-1/uuid`) or plain UUIDs.
- `private_network_ids` - (Optional) List of Private Network IDs to filter the servers by. Accepts either zonal IDs (e.g. `fr-par-1/uuid`) or plain UUIDs.
- `mac_addresses` - (Optional) List of MAC addresses to filter the servers by.

## Attributes Reference

Each result corresponds to one Instance Server and exposes the same attributes as the [`scaleway_instance_server` resource](../resources/instance_server.md).
