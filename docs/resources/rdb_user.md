---
page_title: "scaleway_rdb_user Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_rdb_user`





## Schema

### Required

- **instance_id** (String) Instance on which the user is created
- **name** (String) Database user name
- **password** (String, Sensitive) Database user password

### Optional

- **id** (String) The ID of this resource.
- **is_admin** (Boolean) Grant admin permissions to database user
- **region** (String) The region you want to attach the resource to
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


