---
page_title: "scaleway_account_ssh_key Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_account_ssh_key`





## Schema

### Required

- **public_key** (String) The public SSH key

### Optional

- **id** (String) The ID of this resource.
- **name** (String) The name of the SSH key
- **project_id** (String) The project_id you want to attach the resource to
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **organization_id** (String) The organization_id you want to attach the resource to

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


