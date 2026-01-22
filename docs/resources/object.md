---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object"
---

# Resource: scaleway_object

The `scaleway_object` resource allows you to create and manage objects for [Scaleway Object storage](https://www.scaleway.com/en/docs/object-storage/).

Refer to the [dedicated documentation](https://www.scaleway.com/en/docs/object-storage/how-to/upload-files-into-a-bucket/) for more information on Object Storage objects.


## Example Usage

```terraform
### Basic object creation

resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
}

resource "scaleway_object" "some_file" {
  bucket = scaleway_object_bucket.some_bucket.id
  key    = "object_path"

  file = "myfile"
  hash = filemd5("myfile")
}
```

```terraform
### Using Write-Only SSE Customer Key

resource "scaleway_object_bucket" "encrypted_bucket" {
  name = "encrypted-bucket"
}

# Generate an ephemeral encryption key (not stored in the state)
ephemeral "random_password" "encryption_key" {
  length      = 32
  special     = false
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  # Only hex characters for SSE-C keys
  override_special = ""
}

resource "scaleway_object" "encrypted_file" {
  bucket  = scaleway_object_bucket.encrypted_bucket.id
  key     = "secret-file"
  content = "This is a secret content"

  # Use write-only encryption key
  sse_customer_key_wo         = ephemeral.random_password.encryption_key.result
  sse_customer_key_wo_version = 1
}
```



## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket, or its Terraform ID.

* `key` - (Required) The path to the object.

* `file` - (Optional) The name of the file to upload, defaults to an empty file.

* `content` - (Optional) The content of the file to upload. Only one of `file`, `content` or `content_base64` can be defined.

* `content_base64` - (Optional) The base64-encoded content of the file to upload. Only one of `file`, `content` or `content_base64` can be defined.

-> **Note:** Only one of `file`, `content` or `content_base64` can be defined.

* `hash` - (Optional) Hash of the file, used to trigger the upload on file change.

* `storage_class` - (Optional) Specifies the Scaleway [storage class](https://www.scaleway.com/en/docs/object-storage/concepts/#storage-class) (`STANDARD`, `GLACIER`, or `ONEZONE_IA`) used to store the object.

* `visibility` - (Optional) Visibility of the object, `public-read` or `private`.

* `metadata` - (Optional) Map of metadata used for the object (keys must be lowercase).

* `content_type` -  (Optional) The standard MIME type of the object's content (e.g., 'application/json', 'text/plain'). This specifies how the object should be interpreted by clients. See RFC 9110: <https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type>

* `tags` - (Optional) Map of tags.

* `sse_customer_key` - (Optional) Customer's encryption keys to encrypt data (SSE-C). Only one of `sse_customer_key` or `sse_customer_key_wo` should be specified.

* `sse_customer_key_wo` - (Optional) Customer's encryption keys to encrypt data (SSE-C) in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `sse_customer_key` or `sse_customer_key_wo` should be specified. `sse_customer_key_wo` will not be set in the Terraform state. To update the `sse_customer_key_wo`, you must also update the `sse_customer_key_wo_version`. **Important:** Objects encrypted with `sse_customer_key_wo` cannot be read back by Terraform since the encryption key is not stored in state. This means attributes like `content_type`, `metadata`, and `tags` will not be populated from the actual object.

* `sse_customer_key_wo_version` - (Optional) The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) SSE customer key. To update the `sse_customer_key_wo`, you must also update the `sse_customer_key_wo_version`.

* `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project the bucket is associated with.

~> **Important:** The `project_id` attribute has a particular behavior with s3 products because the s3 API is scoped by project.
If you are using a project different from the default one, you have to specify the `project_id` for every child resource of the bucket,
like objects. Otherwise, Terraform will try to create the child resource with the default project ID and you will get a 403 error.

## Attributes Reference

The `scaleway_object` resource exports certain attributes once the object is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

* `id` - The path of the object, including the name of the bucket.

~> **Important:** Object IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{bucket-name}/{key}`, e.g. `fr-par/bucket-name/object-key`.

* `region` - The Scaleway [region](../guides/regions_and_zones.md) the bucket resides in.

## Import

Objects can be imported using the `{region}/{bucketName}/{objectKey}` identifier, as shown below:

```bash
terraform import scaleway_object.some_object fr-par/some-bucket/some-file
```

~> **Important:** The `project_id` attribute has a particular behavior with s3 products because the s3 API is scoped by project.
If you are using a project different from the default one, you have to specify the project ID at the end of the import command.

```bash
terraform import scaleway_object.some_object fr-par/some-bucket/some-file@xxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx
```
