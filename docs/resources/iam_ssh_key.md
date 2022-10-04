---
page_title: "Scaleway: scaleway_iam_ssh_key"
description: |-
Manages Scaleway IAM SSH Keys.
---

# scaleway_iam_ssh_key

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Creates and manages Scaleway IAM SSH Keys.
For more information, see [the documentation](https://developers.scaleway.com/en/products/iam/api/v1alpha1/#ssh-keys-d8ccd4).

## Example Usage

```hcl
resource "scaleway_iam_ssh_key" "main" {
    name       = "main"
    public_key = "<YOUR-PUBLIC-SSH-KEY>"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the SSH key.
- `public_key` - (Required) The public SSH key to be added.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the SSH key is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the SSH public key.
- `public_key` - The SSH public key string
- `organization_id` - The ID of the organization the SSH key is associated with.
- `created_at` - The date and time of the creation of the SSH key.
- `updated_at` - The date and time of the last update of the SSH key.
- `disabled` - The SSH key status.

## Import

SSH keys can be imported using the `id`, e.g.

```bash
$ terraform import scaleway_iam_ssh_key.main 11111111-1111-1111-1111-111111111111
```
