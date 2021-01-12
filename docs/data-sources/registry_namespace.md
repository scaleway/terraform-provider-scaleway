---
page_title: "scaleway_registry_namespace Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_registry_namespace`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String) The name of the container registry namespace
- **namespace_id** (String) The ID of the registry namespace
- **region** (String) The region you want to attach the resource to

### Read-only

- **description** (String) The description of the container registry namespace
- **endpoint** (String) The endpoint reachable by docker
- **is_public** (Boolean) Define the default visibity policy
- **organization_id** (String) The organization_id you want to attach the resource to
- **project_id** (String) The project_id you want to attach the resource to


