---
page_title: "scaleway_registry_namespace Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_registry_namespace`





## Schema

### Required

- **name** (String) The name of the container registry namespace

### Optional

- **description** (String) The description of the container registry namespace
- **id** (String) The ID of this resource.
- **is_public** (Boolean) Define the default visibity policy
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **endpoint** (String) The endpoint reachable by docker
- **organization_id** (String) The organization_id you want to attach the resource to

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


