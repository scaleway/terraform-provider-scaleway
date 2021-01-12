---
page_title: "scaleway_instance_server Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_instance_server`





## Schema

### Required

- **image** (String) The UUID or the label of the base image used by the server
- **type** (String) The instance type of the server

### Optional

- **additional_volume_ids** (List of String) The additional volumes attached to the server
- **boot_type** (String) The boot type of the server
- **bootscript_id** (String) ID of the target bootscript (set boot_type to bootscript)
- **cloud_init** (String) The cloud init script associated with this server
- **enable_dynamic_ip** (Boolean) Enable dynamic IP on the server
- **enable_ipv6** (Boolean) Determines if IPv6 is enabled for the server
- **id** (String) The ID of this resource.
- **ip_id** (String) The ID of the reserved IP for the server
- **name** (String) The name of the server
- **placement_group_id** (String) The placement group the server is attached to
- **project_id** (String) The project_id you want to attach the resource to
- **root_volume** (Block List, Max: 1) Root volume attached to the server on creation (see [below for nested schema](#nestedblock--root_volume))
- **security_group_id** (String) The security group the server is attached to
- **state** (String) The state of the server should be: started, stopped, standby
- **tags** (List of String) The tags associated with the server
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **user_data** (Block Set, Max: 98) The user data associated with the server (see [below for nested schema](#nestedblock--user_data))
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **ipv6_address** (String) The default public IPv6 address routed to the server.
- **ipv6_gateway** (String) The IPv6 gateway address
- **ipv6_prefix_length** (Number) The IPv6 prefix length routed to the server.
- **organization_id** (String) The organization_id you want to attach the resource to
- **placement_group_policy_respected** (Boolean) True when the placement group policy is respected
- **private_ip** (String) The Scaleway internal IP address of the server
- **public_ip** (String) The public IPv4 address of the server

<a id="nestedblock--root_volume"></a>
### Nested Schema for `root_volume`

Optional:

- **delete_on_termination** (Boolean) Force deletion of the root volume on instance termination
- **size_in_gb** (Number) Size of the root volume in gigabytes

Read-only:

- **volume_id** (String) Volume ID of the root volume


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedblock--user_data"></a>
### Nested Schema for `user_data`

Required:

- **key** (String) A user data key, the value "cloud-init" is not allowed
- **value** (String) A user value


