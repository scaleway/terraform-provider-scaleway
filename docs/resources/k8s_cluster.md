---
page_title: "scaleway_k8s_cluster Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_k8s_cluster`





## Schema

### Required

- **cni** (String) The CNI plugin of the cluster
- **name** (String) The name of the cluster
- **version** (String) The version of the cluster

### Optional

- **admission_plugins** (List of String) The list of admission plugins to enable on the cluster
- **auto_upgrade** (Block List, Max: 1) The auto upgrade configuration for the cluster (see [below for nested schema](#nestedblock--auto_upgrade))
- **autoscaler_config** (Block List, Max: 1) The autoscaler configuration for the cluster (see [below for nested schema](#nestedblock--autoscaler_config))
- **delete_additional_resources** (Boolean) Delete additional resources like block volumes and loadbalancers on cluster deletion
- **description** (String) The description of the cluster
- **enable_dashboard** (Boolean) Enable the dashboard on the cluster
- **feature_gates** (List of String) The list of feature gates to enable on the cluster
- **id** (String) The ID of this resource.
- **ingress** (String) The ingress to be deployed on the cluster
- **project_id** (String) The project_id you want to attach the resource to
- **region** (String) The region you want to attach the resource to
- **tags** (List of String) The tags associated with the cluster
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **apiserver_url** (String) Kubernetes API server URL
- **created_at** (String) The date and time of the creation of the Kubernetes cluster
- **kubeconfig** (List of Object) The kubeconfig configuration file of the Kubernetes cluster (see [below for nested schema](#nestedatt--kubeconfig))
- **organization_id** (String) The organization_id you want to attach the resource to
- **status** (String) The status of the cluster
- **updated_at** (String) The date and time of the last update of the Kubernetes cluster
- **upgrade_available** (Boolean) True if an upgrade is available
- **wildcard_dns** (String) Wildcard DNS pointing to all the ready nodes

<a id="nestedblock--auto_upgrade"></a>
### Nested Schema for `auto_upgrade`

Required:

- **enable** (Boolean) Enables the Kubernetes patch version auto upgrade
- **maintenance_window_day** (String) Day of the maintenance window
- **maintenance_window_start_hour** (Number) Start hour of the 2-hour maintenance window


<a id="nestedblock--autoscaler_config"></a>
### Nested Schema for `autoscaler_config`

Optional:

- **balance_similar_node_groups** (Boolean) Detect similar node groups and balance the number of nodes between them
- **disable_scale_down** (Boolean) Disable the scale down feature of the autoscaler
- **estimator** (String) Type of resource estimator to be used in scale up
- **expander** (String) Type of node group expander to be used in scale up
- **expendable_pods_priority_cutoff** (Number) Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable
- **ignore_daemonsets_utilization** (Boolean) Ignore DaemonSet pods when calculating resource utilization for scaling down
- **scale_down_delay_after_add** (String) How long after scale up that scale down evaluation resumes
- **scale_down_unneeded_time** (String) How long a node should be unneeded before it is eligible for scale down


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


<a id="nestedatt--kubeconfig"></a>
### Nested Schema for `kubeconfig`

Read-only:

- **cluster_ca_certificate** (String)
- **config_file** (String)
- **host** (String)
- **token** (String)


