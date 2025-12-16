Creates and manages Scaleway Kubernetes Cluster authorized IPs.
For more information, please refer to the [API documentation](https://www.scaleway.com/en/developers/api/kubernetes/#path-access-control-list-add-new-acls).

~> **Important:** When creating a Cluster, it comes with a default ACL rule allowing all ranges `0.0.0.0/0`.
Defining custom ACLs with Terraform will overwrite this rule, but it will be recreated automatically when deleting the ACL resource.
