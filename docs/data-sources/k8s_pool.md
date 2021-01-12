---
page_title: "scaleway_k8s_pool Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_k8s_pool`





## Schema

### Optional

- **cluster_id** (String) The ID of the cluster on which this pool will be created
- **id** (String) The ID of this resource.
- **name** (String) The name of the cluster
- **pool_id** (String) The ID of the pool
- **region** (String) The region you want to attach the resource to
- **size** (Number) Size of the pool

### Read-only

- **autohealing** (Boolean) Enable the autohealing on the pool
- **autoscaling** (Boolean) Enable the autoscaling on the pool
- **container_runtime** (String) Container runtime for the pool
- **created_at** (String) The date and time of the creation of the pool
- **current_size** (Number) The actual size of the pool
- **max_size** (Number) Maximum size of the pool
- **min_size** (Number) Minimun size of the pool
- **node_type** (String) Server type of the pool servers
- **nodes** (List of Object) (see [below for nested schema](#nestedatt--nodes))
- **placement_group_id** (String) ID of the placement group
- **status** (String) The status of the pool
- **tags** (List of String) The tags associated with the pool
- **updated_at** (String) The date and time of the last update of the pool
- **version** (String) The Kubernetes version of the pool
- **wait_for_pool_ready** (Boolean) Whether to wait for the pool to be ready

<a id="nestedatt--nodes"></a>
### Nested Schema for `nodes`

Read-only:

- **name** (String)
- **public_ip** (String)
- **public_ip_v6** (String)
- **status** (String)


