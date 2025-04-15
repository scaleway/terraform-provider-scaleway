---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_tls_stage"
---

# Resource: scaleway_edge_services_tls_stage

Creates and manages Scaleway Edge Services TLS Stages.

## Example Usage

### Managed

```terraform
resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id         = scaleway_edge_services_pipeline.main.id
  managed_certificate = true
}
```

### With a certificate stored in Scaleway Secret Manager

```terraform
resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id     = scaleway_edge_services_pipeline.main.id
  secrets {
    secret_id = "11111111-1111-1111-1111-111111111111"
    region    = "fr-par"
  }
}
```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `backend_stage_id` - (Optional) The backend stage ID the TLS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id`, `route_stage_id` and `waf_stage_id` should be specified.
- `cache_stage_id` - (Optional) The cache stage ID the TLS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id`, `route_stage_id` and `waf_stage_id` should be specified.
- `route_stage_id` - (Optional) The route stage ID the TLS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id`, `route_stage_id` and `waf_stage_id` should be specified.
- `waf_stage_id` - (Optional) The WAF stage ID the TLS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id`, `route_stage_id` and `waf_stage_id` should be specified.
- `managed_certificate` - (Optional) Set to true when Scaleway generates and manages a Let's Encrypt certificate for the TLS stage/custom endpoint.
- `secrets` - (Optional) The TLS secrets.
    - `bucket_name` - The ID of the secret.
    - `region` - The region of the secret.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the TLS stage is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the TLS stage (UUID format).
- `certificate_expires_at` - The expiration date of the certificate.
- `created_at` - The date and time of the creation of the TLS stage.
- `updated_at` - The date and time of the last update of the TLS stage.

## Import

TLS stages can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_edge_services_tls_stage.basic 11111111-1111-1111-1111-111111111111
```
