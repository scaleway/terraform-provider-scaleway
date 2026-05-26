---
page_title: "scaleway_object_bucket_server_side_encryption_configuration Data Source - terraform-provider-scaleway"
subcategory: "Object"
description: |-
  Get information about a bucket server side encryption configuration.
---

# scaleway_object_bucket_server_side_encryption_configuration (Data Source)

Get information about a bucket server side encryption configuration. This data source allows you to retrieve information about the server-side encryption configuration of a bucket.




## Example Usage

```terraform
# Get by ID
data "scaleway_object_bucket_server_side_encryption_configuration" "by_id" {
  bucket_server_side_encryption_configuration_id = scaleway_object_bucket_server_side_encryption_configuration.main.id
}
```

```terraform
# Get by Bucket Name
data "scaleway_object_bucket_server_side_encryption_configuration" "by_bucket" {
  bucket     = scaleway_object_bucket.main.name
  project_id = scaleway_object_bucket.main.project_id
}
```




## Arguments Reference

- `bucket_server_side_encryption_configuration_id` - (Optional, String) The ID of the bucket server side encryption configuration. Conflicts with `bucket`.

- `bucket` - (Optional, String) The bucket's name or regional ID. Conflicts with `bucket_server_side_encryption_configuration_id`.

- `project_id` - (Optional, String) The ID of the project the bucket is associated with.

~> **Important:** The `project_id` attribute has a particular behavior with S3 products because the S3 API is scoped by project.
If you are using a project different from the default one, you have to specify the `project_id` when reading the data source.
Otherwise, Terraform will use the default project ID and you will get a 403 error.

## Attributes Reference

- `bucket` - (String) The bucket's name or regional ID.

- `project_id` - (String) The ID of the project the bucket is associated with.

- `rule` - (Set of Object) Set of server-side encryption configuration rules.
    - `apply_server_side_encryption_by_default` - (List of Object) Single object for setting server-side encryption by default.
        - `sse_algorithm` - (String) Server-side encryption algorithm to use. Valid values are AES256.
