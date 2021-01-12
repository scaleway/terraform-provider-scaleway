---
page_title: "scaleway_object_bucket Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_object_bucket`





## Schema

### Required

- **name** (String) The name of the bucket

### Optional

- **acl** (String) ACL of the bucket: either 'public-read' or 'private'.
- **cors_rule** (Block List) (see [below for nested schema](#nestedblock--cors_rule))
- **id** (String) The ID of this resource.
- **region** (String) The region you want to attach the resource to
- **tags** (Map of String) The tags associated with this bucket
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **versioning** (Block List, Max: 1) (see [below for nested schema](#nestedblock--versioning))

### Read-only

- **endpoint** (String) Endpoint of the bucket

<a id="nestedblock--cors_rule"></a>
### Nested Schema for `cors_rule`

Required:

- **allowed_methods** (List of String)
- **allowed_origins** (List of String)

Optional:

- **allowed_headers** (List of String)
- **expose_headers** (List of String)
- **max_age_seconds** (Number)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedblock--versioning"></a>
### Nested Schema for `versioning`

Optional:

- **enabled** (Boolean)


