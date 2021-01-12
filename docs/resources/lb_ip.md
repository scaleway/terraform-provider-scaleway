---
page_title: "scaleway_lb_ip Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_lb_ip`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **reverse** (String) The reverse domain name for this IP
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **ip_address** (String) The load-balancer public IP address
- **lb_id** (String) The ID of the loadbalancer attached to this IP, if any
- **organization_id** (String) The organization_id you want to attach the resource to

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


