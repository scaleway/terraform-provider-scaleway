---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc"
---

# scaleway_vpc

Gets information about a Scaleway Virtual Private Cloud.

## Example Usage

```hcl
# Get info by name
data "scaleway_vpc" "by_name" {
  name = "foobar"
}

# Get info by ID
data "scaleway_vpc" "by_id" {
  vpc_id = "11111111-1111-1111-1111-111111111111"
}

# Get default VPC info
data "scaleway_vpc" "default" {
  is_default = true
}
```

## Argument Reference

* `name` - (Optional) Name of the VPC. A maximum of one of `name` and `vpc_id` should be specified.
* `vpc_id` - (Optional) ID of the VPC. A maximum of one of `name` and `vpc_id` should be specified.
* `is_default` - (Optional) Whether the targeted VPC is the default VPC.
* `organization_id` - The ID of the Organization the VPC is associated with.
* `project_id` - (Optional. Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the VPC is associated with.

## Attributes Reference

Exported attributes are the ones from `scaleway_vpc` [resource](../resources/vpc.md)
