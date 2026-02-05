---
subcategory: "OpenSearch"
page_title: "Scaleway: scaleway_opensearch_deployment"
---

# Resource: scaleway_opensearch_deployment

Creates and manages Scaleway OpenSearch deployments.
For more information refer to the [product documentation](https://www.scaleway.com/en/docs/managed-opensearch/).

## Example Usage

### Basic

```terraform
resource "scaleway_opensearch_deployment" "main" {
  name        = "my-opensearch-cluster"
  version     = "2.0"
  node_amount = 1
  node_type   = "SEARCHDB-SHARED-2C-8G"
  password    = "ThisIsASecurePassword123!"
  
  volume {
    type       = "sbs_5k"
    size_bytes = 5000000000
  }
}
```

### Production Setup with High Availability

```terraform
resource "scaleway_opensearch_deployment" "prod" {
  name        = "logs-prod-cluster"
  version     = "2.0"
  node_amount = 3  # High availability with 3 nodes
  node_type   = "SEARCHDB-DEDICATED-2C-8G"
  password    = var.opensearch_password
  tags        = ["production", "logs"]
  
  volume {
    type       = "sbs_15k"  # High IOPS for production
    size_bytes = 100000000000  # 100 GB
  }
}

output "opensearch_url" {
  value     = scaleway_opensearch_deployment.prod.endpoints[0].services[0].url
  sensitive = false
}
```

### With Tags for Organization

```terraform
resource "scaleway_opensearch_deployment" "analytics" {
  name        = "analytics-cluster"
  version     = "2.0"
  node_amount = 1
  node_type   = "SEARCHDB-SHARED-4C-16G"
  password    = var.opensearch_password
  tags        = ["analytics", "dev", "team-data"]
  
  volume {
    type       = "sbs_5k"
    size_bytes = 10000000000
  }
}
```

## Argument Reference

The following arguments are supported:

- `version` - (Required, Forces new resource) OpenSearch version to use (e.g., "2.0"). Changing this forces recreation of the deployment.
- `node_amount` - (Required) Number of nodes in the cluster. Can be upgraded to add more nodes for high availability.
- `node_type` - (Required, Forces new resource) Type of node to use (e.g., "SEARCHDB-SHARED-2C-8G", "SEARCHDB-DEDICATED-2C-8G"). Changing this forces recreation of the deployment.
- `volume` - (Required) Volume configuration for the cluster.
    - `type` - (Required, Forces new resource) Volume type. Valid values are `sbs_5k` (5K IOPS) or `sbs_15k` (15K IOPS). Changing this forces recreation of the deployment.
    - `size_bytes` - (Required) Volume size in bytes. Can be increased but not decreased.
- `password` - (Optional, Forces new resource) Password for the OpenSearch user. Must be at least 12 characters long. If not specified, you will need to reset the password through the API or console. Changing this forces recreation of the deployment.
- `name` - (Optional) Name of the OpenSearch deployment. If not specified, a random name will be generated.
- `user_name` - (Optional, Forces new resource) Username for the deployment. If not specified, the default username will be used. Changing this forces recreation of the deployment.
- `tags` - (Optional) List of tags to apply to the deployment.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the deployment is associated with.

~> **Important:** A public endpoint is automatically created by default. Private network endpoints can be added using a separate endpoint resource (coming soon).

~> **Important:** The password must be at least 12 characters long. If not provided, you will need to reset it through the Scaleway console or API.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the deployment in the format `{region}/{id}`.
- `status` - The status of the deployment (e.g., "ready", "creating", "upgrading").
- `created_at` - Date and time of deployment creation (RFC 3339 format).
- `updated_at` - Date and time of deployment last update (RFC 3339 format).
- `endpoints` - List of endpoints for accessing the deployment.
    - `id` - The ID of the endpoint.
    - `services` - List of services exposed on the endpoint.
        - `name` - Service name (e.g., "opensearch", "dashboards").
        - `port` - Service port number.
        - `url` - Full URL to access the service (e.g., "https://abc-123.searchdb.fr-par.scw.cloud:9200").
    - `public` - Whether the endpoint is public (true) or private (false).
    - `private_network_id` - Private network ID if the endpoint is private.

## Upgrade Notes

### Scaling the Cluster

You can scale your OpenSearch cluster by modifying the following attributes:

- `node_amount` - Add more nodes for high availability and better performance.
- `volume.size_bytes` - Increase storage capacity (cannot be decreased).

These operations are performed in-place without recreating the cluster.

### Changing Node Type

Changing the `node_type` requires recreating the deployment. Plan accordingly:

1. Create a snapshot of your data (manual process)
2. Modify the `node_type` in your Terraform configuration
3. Apply the changes (will destroy and recreate)
4. Restore your data from the snapshot

## Import

OpenSearch deployments can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_opensearch_deployment.main fr-par/11111111-1111-1111-1111-111111111111
```
