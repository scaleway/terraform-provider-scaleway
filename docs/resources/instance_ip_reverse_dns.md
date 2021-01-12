---
page_title: "scaleway_instance_ip_reverse_dns Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_instance_ip_reverse_dns`





## Schema

### Required

- **ip_id** (String) The IP ID or IP address
- **reverse** (String) The reverse DNS for this IP

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **zone** (String) The zone you want to attach the resource to

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


