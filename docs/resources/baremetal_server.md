---
page_title: "scaleway_baremetal_server Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_baremetal_server`





## Schema

### Required

- **offer** (String) ID or name of the server offer
- **os** (String) The base image of the server
- **ssh_key_ids** (List of String) Array of SSH key IDs allowed to SSH to the server

### Optional

- **description** (String) Some description to associate to the server, max 255 characters
- **hostname** (String) Hostname of the server
- **id** (String) The ID of this resource.
- **name** (String) Name of the server
- **project_id** (String) The project_id you want to attach the resource to
- **tags** (List of String) Array of tags to associate with the server
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **domain** (String)
- **ips** (List of Object) (see [below for nested schema](#nestedatt--ips))
- **offer_id** (String) ID of the server offer
- **organization_id** (String) The organization_id you want to attach the resource to
- **os_id** (String) The base image ID of the server

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedatt--ips"></a>
### Nested Schema for `ips`

Read-only:

- **address** (String)
- **id** (String)
- **reverse** (String)
- **version** (String)


