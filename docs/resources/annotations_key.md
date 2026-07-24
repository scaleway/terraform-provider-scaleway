---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_key"
---

# Resource: scaleway_annotations_key
Create an annotation key to define custom metadata labels that can be attached to Scaleway resources. Annotation keys allow you to organize and categorize resources using custom tags with meaningful names and descriptions.



## Example Usage

```terraform
resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}
```



## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the annotation key.
- `description` - (Optional) Description of the annotation key.
- `organization_id` - (Defaults to [provider](../index.md#arguments-reference) `organization_id`) The [organization ID](../guides/organizations.md) to create the key in.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the annotation key.

## Import

Annotation keys can be imported using their `id`:

```bash
terraform import scaleway_annotations_key.main <key_id>
```
