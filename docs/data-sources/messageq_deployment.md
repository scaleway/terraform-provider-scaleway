---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_deployment"
---

# Data Source: scaleway_messageq_deployment

Gets information about a Scaleway MessageQ deployment.

## Example Usage

```terraform
resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-SHARED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  volume = {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

data "scaleway_messageq_deployment" "by_name" {
  name = scaleway_messageq_deployment.main.name
}
```

## Argument Reference

- `name` - (Optional) The name of the deployment. Only one of `name` and `deployment_id` should be specified.
- `deployment_id` - (Optional) The ID of the deployment. Only one of `name` and `deployment_id` should be specified.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The region in which the deployment exists.
- `project_id` - (Optional) The ID of the project the deployment is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the deployment.
- `name` - Name of the deployment.
- `project_id` - Project ID the deployment belongs to.
- `tags` - Tags of the deployment.
- `version` - MessageQ version.
- `node_count` - Number of nodes.
- `node_type` - Type of node.
- `volume` - Volume configuration (`type`, `size_in_gb`).
- `endpoints` - All endpoints returned by the API for the deployment (public and private when both exist).
    - `id` - The ID of the endpoint.
    - `services` - List of services exposed on the endpoint.
        - `name` - Service name.
        - `port` - Service port.
        - `url` - Service URL.
    - `public` - Whether the endpoint is public (`true`) or private (`false`).
    - `private_network_id` - Private network ID if the endpoint is private.
- `status` - Status of the deployment.
- `created_at` - Creation date (RFC 3339).
- `updated_at` - Last update date (RFC 3339).
