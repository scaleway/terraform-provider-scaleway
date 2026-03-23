---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_server_side_encryption_configuration"
---

# Resource: scaleway_object_bucket_server_side_encryption_configuration

The `scaleway_object_bucket_server_side_encryption_configuration` resource allows you to manage server-side encryption configuration for [Scaleway Object Storage](https://www.scaleway.com/en/docs/object-storage/) buckets.

Refer to the [dedicated documentation](https://www.scaleway.com/en/docs/object-storage/api-cli/enable-sse-one/) for more information on server-side encryption.


## Example Usage

```terraform
resource "scaleway_object_bucket" "test" {
  name   = "my-unique-bucket-name"
  region = "fr-par"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "fr-par"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
```
```terraform
resource "scaleway_object_bucket" "test" {
  name = "my-unique-bucket-name"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The bucket's name or regional ID.

* `rule` - (Required) Set of server-side encryption configuration rules. The `rule` object supports the following:
    * `apply_server_side_encryption_by_default` - (Optional) Single object for setting server-side encryption by default. The `apply_server_side_encryption_by_default` object supports the following:
        * `sse_algorithm` - (Required) Server-side encryption algorithm to use. Valid values are `AES256`.

* `region` - (Optional) The [region](https://www.scaleway.com/en/developers/api/#region-definition) in which the bucket is located.

## Attributes Reference

The `scaleway_object_bucket_server_side_encryption_configuration` resource exports certain attributes once the configuration is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

* `id` - The unique identifier of the server-side encryption configuration.

~> **Important:** Server-side encryption configuration IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{bucket-name}`, e.g. `fr-par/bucket-name`.

* `region` - The Scaleway [region](../guides/regions_and_zones.md) the bucket resides in.

## Import

Server-side encryption configurations can be imported using the `{region}/{bucketName}` identifier, as shown below:

```bash
terraform import scaleway_object_bucket_server_side_encryption_configuration.test fr-par/my-bucket-name
```

~> **Important:** The `project_id` attribute has a particular behavior with S3 products because the S3 API is scoped by project.
If you are using a project different from the default one, you have to specify the project ID at the end of the import command.

```bash
terraform import scaleway_object_bucket_server_side_encryption_configuration.test fr-par/my-bucket-name@11111111-1111-1111-1111-111111111111
```
