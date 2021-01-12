---
page_title: "scaleway_instance_ip Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_instance_ip`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **project_id** (String) The project_id you want to attach the resource to
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **address** (String) The IP address
- **organization_id** (String) The organization_id you want to attach the resource to
- **reverse** (String) The reverse DNS for this IP
- **server_id** (String) The server associated with this IP

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


