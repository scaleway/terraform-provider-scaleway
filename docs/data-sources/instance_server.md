---
page_title: "scaleway_instance_server Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_instance_server`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String) The name of the server
- **server_id** (String) The ID of the server
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **additional_volume_ids** (List of String) The additional volumes attached to the server
- **boot_type** (String) The boot type of the server
- **bootscript_id** (String) ID of the target bootscript (set boot_type to bootscript)
- **cloud_init** (String) The cloud init script associated with this server
- **enable_dynamic_ip** (Boolean) Enable dynamic IP on the server
- **enable_ipv6** (Boolean) Determines if IPv6 is enabled for the server
- **image** (String) The UUID or the label of the base image used by the server
- **ip_id** (String) The ID of the reserved IP for the server
- **ipv6_address** (String) The default public IPv6 address routed to the server.
- **ipv6_gateway** (String) The IPv6 gateway address
- **ipv6_prefix_length** (Number) The IPv6 prefix length routed to the server.
- **organization_id** (String) The organization_id you want to attach the resource to
- **placement_group_id** (String) The placement group the server is attached to
- **placement_group_policy_respected** (Boolean) True when the placement group policy is respected
- **private_ip** (String) The Scaleway internal IP address of the server
- **project_id** (String) The project_id you want to attach the resource to
- **public_ip** (String) The public IPv4 address of the server
- **root_volume** (List of Object) Root volume attached to the server on creation (see [below for nested schema](#nestedatt--root_volume))
- **security_group_id** (String) The security group the server is attached to
- **state** (String) The state of the server should be: started, stopped, standby
- **tags** (List of String) The tags associated with the server
- **type** (String) The instance type of the server
- **user_data** (Set of Object) The user data associated with the server (see [below for nested schema](#nestedatt--user_data))

<a id="nestedatt--root_volume"></a>
### Nested Schema for `root_volume`

Read-only:

- **delete_on_termination** (Boolean)
- **size_in_gb** (Number)
- **volume_id** (String)


<a id="nestedatt--user_data"></a>
### Nested Schema for `user_data`

Read-only:

- **key** (String)
- **value** (String)


