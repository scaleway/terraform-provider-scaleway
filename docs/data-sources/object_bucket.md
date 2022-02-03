---
page_title: "Scaleway: scaleway_object_bucket"
description: |-
  Manages Scaleway object storage buckets.
---

# scaleway_object_bucket

Gets information about the Bucket.
For more information, see [the documentation](https://www.scaleway.com/en/docs/object-storage-feature/).

## Example Usage

```hcl
resource "scaleway_object_bucket" "main" {
    name = "bucket.test.com"
    tags = {
        foo = "bar"
    }
}

data "scaleway_object_bucket" "selected" {
  name = "bucket.test.com"
}
```

## Argument Reference

- `name` - (Required) The bucket name.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the Object Storage exists.


## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The unique name of the bucket.
* `endpoint` - The endpoint URL of the bucket