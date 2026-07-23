---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_binding"
---

# Resource: scaleway_annotations_binding
Creates and manages Scaleway Annotations Bindings.



## Example Usage

```terraform
resource "scaleway_annotations_key" "main" {
  name        = "environment"
  description = "Deployment environment"
}

resource "scaleway_annotations_value" "production" {
  key_id      = scaleway_annotations_key.main.id
  name        = "production"
  description = "Production environment"
}

resource "scaleway_key_manager_key" "main" {
  name        = "example-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Example key for binding"
  unprotected = true
}

resource "scaleway_annotations_binding" "main" {
  srn        = scaleway_key_manager_key.main.srn
  value_id   = scaleway_annotations_value.production.id
}
```



## Argument Reference

The following arguments are supported:

- `srn` - (Required) Scaleway Resource Number to associate. Changing this forces a new resource to be created.
- `value_id` - (Required) ID of the value to associate. Changing this forces a new resource to be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the annotation binding.
- `key_id` - ID of the key associated to the binding.

## Import

Annotation bindings can be imported using their `id`:

```bash
terraform import scaleway_annotations_binding.main <binding_id>
```
