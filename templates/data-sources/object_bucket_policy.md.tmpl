---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_policy"
---

# scaleway_object_bucket_policy

The `scaleway_object_bucket_policy` data source is used to retrieve information about the bucket policy of an Object Storage bucket.

Refer to the Object Storage [documentation](https://www.scaleway.com/en/docs/object-storage/api-cli/bucket-policy/) for more information.

## Retrieve the bucket policy of a bucket

The following command allows you to retrieve a bucket policy by its bucket.

```hcl
data "scaleway_object_bucket_policy" "main" {
  bucket = "bucket.test.com"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_object_bucket_policy` data source to filter and retrieve the desired bucket policy. Each argument has a specific purpose:

- `bucket` - (Required) The name of the bucket.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#zones) in which the Object Storage exists.
- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project with which the bucket is associated.

~> **Important:** The `project_id` attribute has a particular behavior with s3 products, because the s3 API is scoped by project.
If you are using a project different from the default one, you have to specify the `project_id` for every child resource of the bucket,
like bucket policies. Otherwise, Terraform will try to create the child resource with the default project ID and you will get a 403 error.

For more information on Object Storage and Scaleway Projects, refer to the [dedicated documentation](https://www.scaleway.com/en/docs/iam/api-cli/using-api-key-object-storage/).

## Attributes Reference

The `scaleway_object_bucket_policy` data source exports certain attributes once the bucket policy information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all above arguments, the following attribute is exported:

* `policy` - The content of the bucket policy in JSON format.
