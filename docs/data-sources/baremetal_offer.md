---
page_title: "scaleway_baremetal_offer Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_baremetal_offer`





## Schema

### Optional

- **id** (String) The ID of this resource.
- **include_disabled** (Boolean) Include disabled offers
- **name** (String) Exact name of the desired offer
- **offer_id** (String) ID of the desired offer
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **bandwidth** (Number) Available Bandwidth with the offer
- **commercial_range** (String) Commercial range of the offer
- **cpu** (List of Object) CPU specifications of the offer (see [below for nested schema](#nestedatt--cpu))
- **disk** (List of Object) Disk specifications of the offer (see [below for nested schema](#nestedatt--disk))
- **memory** (List of Object) Memory specifications of the offer (see [below for nested schema](#nestedatt--memory))
- **stock** (String) Stock status for this offer

<a id="nestedatt--cpu"></a>
### Nested Schema for `cpu`

Read-only:

- **core_count** (Number)
- **frequency** (Number)
- **name** (String)
- **thread_count** (Number)


<a id="nestedatt--disk"></a>
### Nested Schema for `disk`

Read-only:

- **capacity** (Number)
- **type** (String)


<a id="nestedatt--memory"></a>
### Nested Schema for `memory`

Read-only:

- **capacity** (Number)
- **frequency** (Number)
- **is_ecc** (Boolean)
- **type** (String)


