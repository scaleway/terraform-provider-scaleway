---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_key"
---

# Data Source: scaleway_annotations_key

Retrieves information about an existing annotation key using its ID.



## Example Usage

```terraform
resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}

data "scaleway_annotations_key" "main" {
  key_id = scaleway_annotations_key.environment.id
}
```



## Argument Reference

- `key_id` - (Required) The ID of the annotation key to retrieve.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the annotation key
- `name` - Name of the annotation key
- `description` - Description of the annotation key
