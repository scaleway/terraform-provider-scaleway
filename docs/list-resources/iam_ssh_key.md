---
page_title: "Scaleway: scaleway_iam_ssh_key"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM SSH Keys across projects.
---

# Resource: scaleway_iam_ssh_key

Lists Scaleway IAM SSH Keys across projects.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
# List all SSH keys across all projects
list "scaleway_iam_ssh_key" "all" {
  provider = scaleway

  config {
    project_ids = ["*"]
  }
}
```
```terraform
# List SSH keys filtered by name
list "scaleway_iam_ssh_key" "by_name" {
  provider = scaleway

  config {
    name = "my-ssh-key"
  }
}
```
```terraform
# List disabled SSH keys in specific projects
list "scaleway_iam_ssh_key" "disabled" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    disabled = true
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If not specified, the provider default project is used.
- `name` - (Optional) Name of the SSH key to filter for.
- `disabled` - (Optional) Filter SSH keys by disabled status.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each SSH Key:

- `id` - The ID of the SSH Key.
- `name` - The name of the SSH Key.
- `public_key` - The public SSH key.
- `fingerprint` - The fingerprint of the SSH key.
- `created_at` - The date and time of the creation of the SSH Key.
- `updated_at` - The date and time of the last update of the SSH Key.
- `organization_id` - The organization ID the SSH Key belongs to.
- `project_id` - The project ID the SSH Key belongs to.
- `disabled` - The SSH key status.
