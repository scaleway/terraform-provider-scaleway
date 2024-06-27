---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket"
---

# scaleway_object_bucket

The `scaleway_object_bucket` data source is used to retrieve information about an Object Storage bucket.

Refer to the Object Storage [documentation](https://www.scaleway.com/en/docs/storage/object/how-to/create-a-bucket/) for more information.

## Retrieve an Object Storage bucket

The following commands allow you to:

- retrieve a bucket by its name
- retrieve a bucket by its ID

```hcl
resource "scaleway_object_bucket" "main" {
    name = "bucket.test.com"
    tags = {
        foo = "bar"
    }
}

data "scaleway_object_bucket" "selected" {
  name = scaleway_object_bucket.main.id
}
```

## Retrieve a bucket from a specific project

```hcl
data "scaleway_object_bucket" "selected" {
  name = "bucket.test.com"
  project_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_object_bucket` data source to filter and retrieve the desired Object Storage bucket. Each argument has a specific purpose:

- `name` - (Required) The name of the bucket, or its terraform ID (`{region}/{name}`)
- `object_lock_enabled` - (Optional) Enable object lock on the bucket. Defaults to `false`. Updating this field will force the creation of a new bucket.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the bucket exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project with which the bucket is associated.


## Attributes reference

The `scaleway_object_bucket` data source exports certain attributes once the bucket information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all above arguments, the following attribute is exported:

* `id` - The unique identifier of the bucket.

~> **Important:** Object buckets' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{name}`, e.g. `fr-par/bucket-name`

* `endpoint` - The endpoint URL of the bucket