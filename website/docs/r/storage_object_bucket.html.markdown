---
layout: "scaleway"
page_title: "Scaleway: scaleway_storage_object_bucket"
sidebar_current: "docs-scaleway-resource-storage-object-bucket"
description: |-
  Manages Scaleway object storage buckets.
---

# scaleway_storage_object_bucket

Creates and manages Scaleway object storage buckets. For more information, see [the documentation](https://www.scaleway.com/en/docs/object-storage-feature/).

## Example Usage

```hcl
resource "scaleway_storage_object_bucket" "some_bucket" {
    name = "some-unique-name"
    acl = "private"
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name of the bucket.
* `acl` - (Optional) The canned ACL you want to apply to the bucket.
[//]: # (TODO: Add a link to doc when it will exist.)
* `region` - (Optional) The [region](https://developers.scaleway.com/en/quickstart/#region-definition) in which the bucket should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the bucket.



## Import

Buckets can be imported using the `{region}/{id}` identifier, e.g.

```
$ terraform import scaleway_storage_object_bucket.some_bucket fr-par/11111111-1111-1111-1111-111111111111
```
