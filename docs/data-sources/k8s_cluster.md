---
page_title: "scaleway_k8s_cluster Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_k8s_cluster`





## Schema

### Optional

- **cluster_id** (String) The ID of the cluster
- **id** (String) The ID of this resource.
- **name** (String) The name of the cluster
- **region** (String) The region you want to attach the resource to

### Read-only

- **admission_plugins** (List of String) The list of admission plugins to enable on the cluster
- **apiserver_url** (String) Kubernetes API server URL
- **auto_upgrade** (List of Object) The auto upgrade configuration for the cluster (see [below for nested schema](#nestedatt--auto_upgrade))
- **autoscaler_config** (List of Object) The autoscaler configuration for the cluster (see [below for nested schema](#nestedatt--autoscaler_config))
- **cni** (String) The CNI plugin of the cluster
- **created_at** (String) The date and time of the creation of the Kubernetes cluster
- **description** (String) The description of the cluster
- **enable_dashboard** (Boolean) Enable the dashboard on the cluster
- **feature_gates** (List of String) The list of feature gates to enable on the cluster
- **ingress** (String) The ingress to be deployed on the cluster
- **kubeconfig** (List of Object) The kubeconfig configuration file of the Kubernetes cluster (see [below for nested schema](#nestedatt--kubeconfig))
- **organization_id** (String) The organization_id you want to attach the resource to
- **project_id** (String) The project_id you want to attach the resource to
- **status** (String) The status of the cluster
- **tags** (List of String) The tags associated with the cluster
- **updated_at** (String) The date and time of the last update of the Kubernetes cluster
- **upgrade_available** (Boolean) True if an upgrade is available
- **version** (String) The version of the cluster
- **wildcard_dns** (String) Wildcard DNS pointing to all the ready nodes

<a id="nestedatt--auto_upgrade"></a>
### Nested Schema for `auto_upgrade`

Read-only:

- **enable** (Boolean)
- **maintenance_window_day** (String)
- **maintenance_window_start_hour** (Number)


<a id="nestedatt--autoscaler_config"></a>
### Nested Schema for `autoscaler_config`

Read-only:

- **balance_similar_node_groups** (Boolean)
- **disable_scale_down** (Boolean)
- **estimator** (String)
- **expander** (String)
- **expendable_pods_priority_cutoff** (Number)
- **ignore_daemonsets_utilization** (Boolean)
- **scale_down_delay_after_add** (String)
- **scale_down_unneeded_time** (String)


<a id="nestedatt--kubeconfig"></a>
### Nested Schema for `kubeconfig`

Read-only:

- **cluster_ca_certificate** (String)
- **config_file** (String)
- **host** (String)
- **token** (String)


