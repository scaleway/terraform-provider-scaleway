---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_pipeline"
---

# Resource: scaleway_edge_services_pipeline

Creates and manages Scaleway Edge Services Pipelines.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_pipeline" "main" {
  name        = "pipeline-name"
  description = "pipeline description"
}
```

### Complete pipeline

```terraform
resource "scaleway_edge_services_pipeline" "main" {
  name        = "pipeline-name"
  description = "pipeline description"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = "my-bucket-name"
    bucket_region = "fr-par"
  }
}

resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id         = scaleway_edge_services_pipeline.main.id
  cache_stage_id      = scaleway_edge_services_cache_stage.main.id
  managed_certificate = true
}

resource "scaleway_edge_services_dns_stage" "main" {
  pipeline_id  = scaleway_edge_services_pipeline.main.id
  tls_stage_id = scaleway_edge_services_tls_stage.main.id
  fqdns        = ["subdomain.example.com"]
}

resource "scaleway_edge_services_head_stage" "main" {
  pipeline_id   = scaleway_edge_services_pipeline.main.id
  head_stage_id = scaleway_edge_services_dns_stage.main.id
}

resource "scaleway_edge_services_cache_stage" "main" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.main.id
}
```

## Argument Reference

- `name` - (Optional) The name of the pipeline.
- `description` - (Optional) The description of the pipeline.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the pipeline is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the pipeline (UUID format).
- `created_at` - The date and time of the creation of the pipeline.
- `updated_at` - The date and time of the last update of the pipeline.
- `status` - The status of user pipeline.

## Import

Pipelines can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_edge_services_pipeline.basic 11111111-1111-1111-1111-111111111111
```
