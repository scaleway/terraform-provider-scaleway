The [`scaleway_k8s_cluster_acl`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/k8s_cluster_acl) resource allows you to create and manage Scaleway Kubernetes Cluster authorized IPs.

Refer to the Kubernetes [documentation](https://www.scaleway.com/en/docs/compute/kubernetes/) and [API documentation](https://www.scaleway.com/en/developers/api/kubernetes/) for more information.

~> **Important:** When creating a Cluster, it comes with a default ACL rule allowing all ranges `0.0.0.0/0`.
Defining custom ACLs with Terraform will overwrite this rule, but it will be recreated automatically when deleting the ACL resource.
