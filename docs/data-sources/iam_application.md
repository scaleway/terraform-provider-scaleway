---
layout: "scaleway"
page_title: "Scaleway: scaleway_iam_application"
description: |-
Gets information about an existing IAM application.
---

# scaleway_iam_application

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Gets information about an existing IAM application.

## Example Usage

```hcl
# Get info by name
data "scaleway_iam_application" "find_by_id" { 
    name            = "foobar"
    organization_id = "11111111-1111-1111-1111-111111111111"
}
# Get info by application ID
data "scaleway_iam_application" "find_by_name" {
    application_id    = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the IAM application.
  Only one of the `name` and `application_id` should be specified.

- `application_id` - (Optional) The ID of the IAM application.
  Only one of the `name` and `application_id` should be specified.

- `organization_id` - (Optional) The organization ID the IAM group is associated with.

## Attribute Reference

Exported attributes are the ones from `iam_application` [resource](../resources/iam_application.md)
