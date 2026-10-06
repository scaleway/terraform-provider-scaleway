---
subcategory: "Annotations"
page_title: "Scaleway: scaleway_annotations_binding"
---

# Data Source: scaleway_annotations_binding

Use this data source to retrieve information about an annotation binding.



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

resource "scaleway_key_manager_key" "main" {
  name        = "example-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Example key for binding"
  unprotected = true
}

resource "scaleway_annotations_binding" "main" {
  srn      = scaleway_key_manager_key.main.srn
  value_id = scaleway_annotations_value.production.id
}

data "scaleway_annotations_binding" "main" {
  binding_id = scaleway_annotations_binding.main.id
}
```



## Argument Reference

- `binding_id` - (Required) The ID of the annotation binding to retrieve.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the annotation binding
- `srn` - Scaleway Resource Number associated to the binding
- `value_id` - ID of the value associated to the binding
- `key_id` - ID of the key associated to the binding
