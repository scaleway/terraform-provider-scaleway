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

  volume {
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
- All attributes from the [resource](../resources/messageq_deployment.md) are exported.
