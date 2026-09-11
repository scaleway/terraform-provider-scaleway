---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_user"
---

# Resource: scaleway_messageq_user

Creates and manages users on a Scaleway MessageQ deployment.
For more information refer to the [API documentation](https://www.scaleway.com/en/developers/api/messageq).

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

resource "scaleway_messageq_user" "app" {
  deployment_id = scaleway_messageq_deployment.main.id
  name          = "orders-app"
  password      = var.app_password
}
```

## Argument Reference

The following arguments are supported:

- `deployment_id` - (Required, Forces new resource) The ID of the MessageQ deployment.
- `name` - (Required, Forces new resource) The username.
- `password` - (Optional, Sensitive) The user password. Only one of `password` or `password_wo` should be specified.
- `password_wo` - (Optional, Write-only) The user password in [write-only](../guides/using-write-only-arguments.md) mode. Only one of `password` or `password_wo` should be specified. To rotate the password, update both `password_wo` and `password_wo_version`.
- `password_wo_version` - (Optional) Version of the write-only password. Required with `password_wo`.
- `region` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `region`) The region in which the user should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user in the format `{region}/{deployment_id}/{name}`.

## Import

MessageQ users can be imported using `{region}/{deployment_id}/{name}`, e.g.

```bash
terraform import scaleway_messageq_user.app fr-par/11111111-1111-1111-1111-111111111111/orders-app
```
