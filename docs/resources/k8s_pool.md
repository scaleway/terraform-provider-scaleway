---
page_title: "scaleway_k8s_pool Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_k8s_pool`





## Schema

### Required

- **cluster_id** (String) The ID of the cluster on which this pool will be created
- **name** (String) The name of the cluster
- **node_type** (String) Server type of the pool servers
- **size** (Number) Size of the pool

### Optional

- **autohealing** (Boolean) Enable the autohealing on the pool
- **autoscaling** (Boolean) Enable the autoscaling on the pool
- **container_runtime** (String) Container runtime for the pool
- **id** (String) The ID of this resource.
- **max_size** (Number) Maximum size of the pool
- **min_size** (Number) Minimun size of the pool
- **placement_group_id** (String) ID of the placement group
- **region** (String) The region you want to attach the resource to
- **tags** (List of String) The tags associated with the pool
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **wait_for_pool_ready** (Boolean) Whether to wait for the pool to be ready

### Read-only

- **created_at** (String) The date and time of the creation of the pool
- **current_size** (Number) The actual size of the pool
- **nodes** (List of Object) (see [below for nested schema](#nestedatt--nodes))
- **status** (String) The status of the pool
- **updated_at** (String) The date and time of the last update of the pool
- **version** (String) The Kubernetes version of the pool

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedatt--nodes"></a>
### Nested Schema for `nodes`

Read-only:

- **name** (String)
- **public_ip** (String)
- **public_ip_v6** (String)
- **status** (String)


