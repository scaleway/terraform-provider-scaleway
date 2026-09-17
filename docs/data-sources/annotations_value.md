---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_value"
---

# Data Source: scaleway_annotations_value
Retrieves information about an existing annotation value using its ID.



## Example Usage

```terraform
resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}

resource "scaleway_annotations_value" "production" {
  key_id      = scaleway_annotations_key.environment.id
  name        = "production"
  description = "Production environment"
}

data "scaleway_annotations_value" "main" {
  value_id = scaleway_annotations_value.production.id
}
```



## Argument Reference

The following arguments are supported:

- `value_id` - (Required) The ID of the annotation value to retrieve.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the annotation value
- `key_id` - ID of the key the value is associated to
- `name` - Name of the annotation value
- `description` - Description of the annotation value
