---
page_title: "scaleway_lb_frontend Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_lb_frontend`





## Schema

### Required

- **backend_id** (String) The load-balancer backend ID
- **inbound_port** (Number) TCP port to listen on the front side
- **lb_id** (String) The load-balancer ID

### Optional

- **acl** (Block List) ACL rules (see [below for nested schema](#nestedblock--acl))
- **certificate_id** (String) Certificate ID
- **id** (String) The ID of this resource.
- **name** (String) The name of the frontend
- **timeout_client** (String) Set the maximum inactivity time on the client side
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--acl"></a>
### Nested Schema for `acl`

Required:

- **action** (Block List, Min: 1, Max: 1) Action to undertake when an ACL filter matches (see [below for nested schema](#nestedblock--acl--action))
- **match** (Block List, Min: 1, Max: 1) The ACL match rule (see [below for nested schema](#nestedblock--acl--match))

Optional:

- **name** (String) The ACL name
- **region** (String) The region you want to attach the resource to

Read-only:

- **organization_id** (String) The organization_id you want to attach the resource to

<a id="nestedblock--acl--action"></a>
### Nested Schema for `acl.action`

Required:

- **type** (String) The action type


<a id="nestedblock--acl--match"></a>
### Nested Schema for `acl.match`

Optional:

- **http_filter** (String) The HTTP filter to match
- **http_filter_value** (List of String) A list of possible values to match for the given HTTP filter
- **invert** (Boolean) If set to true, the condition will be of type "unless"
- **ip_subnet** (List of String) A list of IPs or CIDR v4/v6 addresses of the client of the session to match



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


