---
page_title: "Scaleway: scaleway_object_bucket"
subcategory: "Object Storage"
description: |-
  Lists Scaleway Object Storage Buckets.
---

# Resource: scaleway_object_bucket

Lists Scaleway Object Storage Buckets.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/storage/object/concepts/).

## Example Usage

```terraform
# List all buckets across all regions and all projects
list "scaleway_object_bucket" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List buckets across all regions, filtered by name prefix
list "scaleway_object_bucket" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-bucket"
  }
}
```

```terraform
# List buckets filtered by project ID
list "scaleway_object_bucket" "by_project" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
// List buckets filtered by region
list "scaleway_object_bucket" "by_region" {
  provider = scaleway

  config {
    regions = ["fr-par"]
  }
}
```

```terraform
# List buckets filtered by tags
list "scaleway_object_bucket" "by_tags" {
  provider = scaleway

  config {
    tags = ["production", "env:prod"]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `project_ids` - (Optional) Project IDs to filter for. Use '*' to list across all projects. If not specified, the provider default project ID is used.
- `regions` - (Optional) Regions to target. Use '*' to list from all regions. If not specified, the provider default region is used.
- `name` - (Optional) Filter by bucket name.
- `tags` - (Optional) Filter by tags.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Bucket:

- `id` - The regional ID of the bucket.
- `name` - The name of the bucket.
- `region` - The region of the bucket.
- `project_id` - The project ID the bucket belongs to.
- `endpoint` - The endpoint URL of the bucket.
- `api_endpoint` - The API endpoint URL of the bucket.

