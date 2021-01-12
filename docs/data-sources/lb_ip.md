---
page_title: "scaleway_lb_ip Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_lb_ip`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **ip_address** (String) The IP address
- **ip_id** (String) The ID of the IP address

### Read-only

- **lb_id** (String) The ID of the loadbalancer attached to this IP, if any
- **organization_id** (String) The organization_id you want to attach the resource to
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **reverse** (String) The reverse domain name for this IP


