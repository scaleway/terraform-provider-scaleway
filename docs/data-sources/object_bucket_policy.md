---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_policy"
---

# scaleway_object_bucket_policy

Gets information about the Bucket's policy.
For more information, see [the documentation](https://www.scaleway.com/en/docs/object-storage-feature/).

## Example Usage

```hcl
data "scaleway_object_bucket_policy" "main" {
    bucket = "bucket.test.com"
}
```

## Argument Reference

- `bucket` - (Required) The bucket name.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the Object Storage exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the bucket is associated with.


## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `policy` - The bucket's policy in JSON format.
