---
subcategory: "Account"
page_title: "Scaleway: scaleway_account_ssh_key"
---

# scaleway_account_ssh_key

The `scaleway_account_ssh_key` data source is used to retrieve information about a the SSH key of a Scaleway account.

Refer to the Organizations and Projects [documentation](https://www.scaleway.com/en/docs/identity-and-access-management/organizations-and-projects/how-to/create-ssh-key/) and [API documentation](https://www.scaleway.com/en/developers/api/iam/#path-ssh-keys) for more information.


## Retrieve the SSH key of a Scaleway account

The following commands allow you to:

- retrieve an SSH key by its name
- retrieve an SSH key by its ID

```hcl
# Get info by SSH key name
data "scaleway_account_ssh_key" "my_key" {
  name  = "my-key-name"
}

# Get info by SSH key id
data "scaleway_account_ssh_key" "my_key" {
  ssh_key_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_account_ssh_key` data source to filter and retrieve the desired SSH key. Each argument has a specific purpose:

- `name` - The name of the SSH key.
- `ssh_key_id` - The unique identifier of the SSH key.

  -> **Note** You must specify at least one: `name` and/or `ssh_key_id`.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The unique identifier of the project with which the SSH key is associated.

## Attributes Reference

The `scaleway_account_ssh_key` data source exports certain attributes once the SSH key information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all above arguments, the following attributes are exported:

- `id` - The unique identifier of the SSH public key.
- `public_key` - The string of the SSH public key.
- `organization_id` - The unique identifier of the Organization with which the SSH key is associated.
