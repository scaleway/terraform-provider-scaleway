---
page_title: "Scaleway: scaleway_object_bucket_lock_configuration"
description: |-
Manages Scaleway lock on object storage buckets.
---

# scaleway_object_bucket_lock_configuration

Provides an Object bucket lock configuration resource.
For more information, see [Setting up object lock](https://www.scaleway.com/en/docs/storage/object/api-cli/object-lock/).

## Example Usage

### Configure an Object Lock for a new bucket

Please note that `object_lock_enabled` must be set to `true` before configuring the lock.

```hcl
resource "scaleway_object_bucket" "main" {
    name = "MyBucket"
    acl  = "public-read"

    object_lock_enabled = true
}

resource "scaleway_object_bucket_lock_configuration" "main" {
    bucket = scaleway_object_bucket.main.name

    rule {
        default_retention {
            mode = "GOVERNANCE"
            days = 1
        }
    }
}
```

### Configure an Object Lock for an existing bucket

You should [contact Scaleway support](https://console.scaleway.com/support/tickets/create) to enable object lock on an existing bucket.

## Attributes Reference

The following arguments are supported:

- `bucket` - (Required, Forces new resource) The name of the bucket.

- `rule` - (Optional) Specifies the Object Lock rule for the specified object.

    - `default_retention` - (Required) The default retention for the lock.

        - `mode` - (Required) The default Object Lock retention mode you want to apply to new objects placed in the specified bucket.

        - `days` - (Optional) The number of days that you want to specify for the default retention period.

        - `years` - (Optional) The number of years that you want to specify for the default retention period.

## Import

Lock configuration Bucket can be imported using the `{region}/{bucketName}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_lock_configuration.some_bucket fr-par/some-bucket
```
