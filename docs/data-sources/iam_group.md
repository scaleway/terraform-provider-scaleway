---
layout: "scaleway"
page_title: "Scaleway: scaleway_iam_group"
description: |-
Gets information about an existing IAM group.
---

# scaleway_iam_group

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Gets information about an existing IAM group. For more information, please check [the documentation](https://developers.scaleway.com/en/products/iam/api/v1alpha1/#applications-83ce5e)

## Example Usage

```hcl
# Get info by name
data "scaleway_iam_group" "find_by_name" { 
  name            = "foobar"
}
# Get info by group ID
data "scaleway_iam_group" "find_by_id" {
  group_id  = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the IAM group.
  Only one of the `name` and `group_id` should be specified.

- `group_id` - (Optional) The ID of the IAM group.
  Only one of the `name` and `group_id` should be specified.

- `organization_id` - (Optional) The organization ID the IAM group is associated with.

## Attribute Reference

Exported attributes are the ones from `iam_group` [resource](../resources/iam_group.md)
