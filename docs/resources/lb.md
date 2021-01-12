---
page_title: "scaleway_lb Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_lb`





## Schema

### Required

- **ip_id** (String) The load-balance public IP ID
- **type** (String) The type of load-balancer you want to create

### Optional

- **id** (String) The ID of this resource.
- **name** (String) Name of the lb
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **tags** (List of String) Array of tags to associate with the load-balancer
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **ip_address** (String) The load-balance public IP address
- **organization_id** (String) The organization_id you want to attach the resource to

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


