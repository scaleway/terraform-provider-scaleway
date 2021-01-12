---
page_title: "scaleway_instance_security_group Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_instance_security_group`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String) The name of the security group
- **security_group_id** (String) The ID of the security group
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **description** (String) The description of the security group
- **enable_default_security** (Boolean) Enable blocking of SMTP on IPv4 and IPv6
- **external_rules** (Boolean)
- **inbound_default_policy** (String) Default inbound traffic policy for this security group
- **inbound_rule** (List of Object) Inbound rules for this security group (see [below for nested schema](#nestedatt--inbound_rule))
- **organization_id** (String) The organization_id you want to attach the resource to
- **outbound_default_policy** (String) Default outbound traffic policy for this security group
- **outbound_rule** (List of Object) Outbound rules for this security group (see [below for nested schema](#nestedatt--outbound_rule))
- **project_id** (String) The project_id you want to attach the resource to
- **stateful** (Boolean) The stateful value of the security group

<a id="nestedatt--inbound_rule"></a>
### Nested Schema for `inbound_rule`

Read-only:

- **action** (String)
- **ip** (String)
- **ip_range** (String)
- **port** (Number)
- **port_range** (String)
- **protocol** (String)


<a id="nestedatt--outbound_rule"></a>
### Nested Schema for `outbound_rule`

Read-only:

- **action** (String)
- **ip** (String)
- **ip_range** (String)
- **port** (Number)
- **port_range** (String)
- **protocol** (String)


