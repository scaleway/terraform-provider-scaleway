---
page_title: "scaleway_instance_placement_group Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_instance_placement_group`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **name** (String) The name of the placement group
- **policy_mode** (String) One of the two policy_mode may be selected: enforced or optional.
- **policy_type** (String) The operating mode is selected by a policy_type
- **project_id** (String) The project_id you want to attach the resource to
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **organization_id** (String) The organization_id you want to attach the resource to
- **policy_respected** (Boolean) Is true when the policy is respected.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


