---
page_title: "Scaleway: scaleway_object_bucket_acl"
description: |-
Manages Scaleway object storage bucket ACL resource.
---

# scaleway_object_bucket

Creates and manages Scaleway object storage bucket ACL.
For more information, see [the documentation](https://www.scaleway.com/en/docs/storage/object/concepts/#access-control-list-(acl)).

-> **Note:** `terraform destroy`  does not delete the Object Bucket ACL but does remove the resource from Terraform state.

-> **Note:** [Account identifiers](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html) is not supported by scaleway.

## Example Usage

```hcl
resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
}

resource "scaleway_object_bucket_acl" "main" {
  bucket = scaleway_object_bucket.main.name
  acl = "private"
}
```

## Example with Grants

```hcl
resource "scaleway_object_bucket" "main" {
    name = "your-bucket"
}

resource "scaleway_object_bucket_acl" "main" {
    bucket = scaleway_object_bucket.main.name
    access_control_policy {
      grant {
        grantee {
            id   = "<project-id>:<project-id>"
            type = "CanonicalUser"
        }
        permission = "FULL_CONTROL"
      }
    
      grant {
        grantee {
          id   = "<project-id>:<project-id>"
          type = "CanonicalUser"
        }
        permission = "WRITE"
      }
    
      owner {
        id = "<project-id>:<project-id>"
      }
    }
}
```

## Arguments Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket.
* `acl` - (Optional) The canned ACL you want to apply to the bucket.
* `access_control_policy` - (Optional, Conflicts with acl) A configuration block that sets the ACL permissions for an object per grantee documented below.
* `expected_bucket_owner` - (Optional, Forces new resource) The project ID of the expected bucket owner.
* `region` - (Optional) The [region](https://developers.scaleway.com/en/quickstart/#region-definition) in which the bucket should be created.

## The ACL

Please check the [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl_overview.html#canned-acl)

## The Access Control policy

The `access_control_policy` configuration block supports the following arguments:

* `grant` - (Required) Set of grant configuration blocks documented below.
* `owner` - (Required) Configuration block of the bucket owner's display name and ID documented below.

## The Grant

The `grant` configuration block supports the following arguments:

* `grantee` - (Required) Configuration block for the person being granted permissions documented below.
* `permission` - (Required) Logging permissions assigned to the grantee for the bucket.

## The owner

The `owner` configuration block supports the following arguments:

* `id` - (Required) The ID of the project owner. Format <project_id>:<project_id>.
* `display_name` - (Optional) The display name of the owner. Format <project_id>:<project_id>.

## the grantee

The `grantee` configuration block supports the following arguments:

* `id` - (Required) The canonical user ID of the grantee.
* `type` - (Required) Type of grantee. Valid values: CanonicalUser.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The `region`,`bucket` and `acl` separated by (`/`).

## Import

Buckets can be imported using the `{region}/{bucketName}/{acl}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_acl.some_bucket fr-par/some-bucket
/private```
