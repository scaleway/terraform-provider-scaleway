---
page_title: "scaleway_rdb_instance Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_rdb_instance`





## Schema

### Required

- **engine** (String) Database's engine version id
- **node_type** (String) The type of database instance you want to create

### Optional

- **disable_backup** (Boolean) Disable automated backup for the database instance
- **id** (String) The ID of this resource.
- **is_ha_cluster** (Boolean) Enable or disable high availability for the database instance
- **name** (String) Name of the database instance
- **password** (String, Sensitive) Password for the first user of the database instance
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **tags** (List of String) List of tags ["tag1", "tag2", ...] attached to a database instance
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **user_name** (String) Identifier for the first user of the database instance

### Read-only

- **certificate** (String) Certificate of the database instance
- **endpoint_ip** (String) Endpoint IP of the database instance
- **endpoint_port** (Number) Endpoint port of the database instance
- **organization_id** (String) The organization_id you want to attach the resource to
- **read_replicas** (List of Object) Read replicas of the database instance (see [below for nested schema](#nestedatt--read_replicas))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedatt--read_replicas"></a>
### Nested Schema for `read_replicas`

Read-only:

- **ip** (String)
- **name** (String)
- **port** (Number)


