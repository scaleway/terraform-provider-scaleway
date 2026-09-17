---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_value"
---

# Resource: scaleway_annotations_value
Create an annotation value to define a specific label instance that can be attached to Scaleway resources. Annotation values are associated with annotation keys and represent concrete tag values (e.g., "production" as a value for an "environment" key).



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
```



## Argument Reference

The following arguments are supported:

- `key_id` - (Required) ID of the key the value is associated to.
- `name` - (Required) Name of the annotation value.
- `description` - (Optional) Description of the annotation value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the annotation value.
- `value_id` - The ID of the annotation value.

## Import

Annotation values can be imported using their `id`:

```bash
terraform import scaleway_annotations_value.main <value_id>
```
