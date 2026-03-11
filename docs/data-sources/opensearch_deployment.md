---
subcategory: "OpenSearch"
page_title: "Scaleway: scaleway_opensearch_deployment"
---

# scaleway_opensearch_deployment

Gets information about an OpenSearch deployment.

For further information refer to the Managed OpenSearch [API documentation](https://developers.scaleway.com/en/products/opensearch/api/).

## Example Usage

```hcl
# Get info by name
data "scaleway_opensearch_deployment" "by_name" {
  name = "my-opensearch-cluster"
}

# Get info by deployment ID
data "scaleway_opensearch_deployment" "by_id" {
  deployment_id = "11111111-1111-1111-1111-111111111111"
}

# Get other attributes
output "opensearch_endpoint" {
  description = "API endpoint of the OpenSearch deployment"
  value       = data.scaleway_opensearch_deployment.by_name.endpoints[0].services[0].url
}
```

## Argument Reference

- `name` - (Optional) The name of the OpenSearch deployment.

- `deployment_id` - (Optional) The ID of the OpenSearch deployment.

  -> **Note** You must specify at least one: `name` or `deployment_id`.

- `project_id` - (Optional) The ID of the project the deployment is in. Can be used to filter deployments when using `name`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the OpenSearch deployment exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the OpenSearch deployment in the format `{region}/{id}`.
- `name` - The name of the OpenSearch deployment.
- `version` - The OpenSearch version running on the deployment.
- `node_type` - The type of nodes in the deployment.
- `node_amount` - The number of nodes in the deployment.
- `tags` - A list of tags attached to the deployment.
- `status` - The current status of the deployment (e.g., `ready`, `creating`, `upgrading`).
- `created_at` - The date and time the deployment was created (RFC 3339 format).
- `updated_at` - The date and time the deployment was last updated (RFC 3339 format).
- `project_id` - The ID of the project the deployment belongs to.
- `volume` - The volume configuration of the deployment.
    - `type` - The volume type (`sbs_5k` or `sbs_15k`).
    - `size_in_gb` - The volume size in GB.
- `endpoints` - The list of endpoints to access the deployment.
    - `id` - The endpoint ID.
    - `public` - Whether the endpoint is public (`true`) or private (`false`).
    - `private_network_id` - The private network ID if the endpoint is private.
    - `services` - The list of services exposed on the endpoint.
        - `name` - The service name (e.g., `api`, `dashboard`).
        - `port` - The service port number.
        - `url` - The full URL to access the service.
