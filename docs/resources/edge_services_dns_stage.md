---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_dns_stage"
---

# Resource: scaleway_edge_services_dns_stage

Creates and manages Scaleway Edge Services DNS Stages.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_dns_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  fqdns       = ["subdomain.example.com"]
}
```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `backend_stage_id` - (Optional) The backend stage ID the DNS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id` and `tls_stage_id` should be specified.
- `tls_stage_id` - (Optional) The TLS stage ID the DNS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id` and `tls_stage_id` should be specified.
- `cache_stage_id` - (Optional) The cache stage ID the DNS stage will be linked to. Only one of `backend_stage_id`, `cache_stage_id` and `tls_stage_id` should be specified.
- `fqdns` - (Optional) Fully Qualified Domain Name (in the format subdomain.example.com) to attach to the stage.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the DNS stage is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the DNS stage (UUID format).
- `type` - The type of the stage.
- `created_at` - The date and time of the creation of the DNS stage.
- `updated_at` - The date and time of the last update of the DNS stage.

## Import

DNS stages can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_edge_services_dns_stage.basic 11111111-1111-1111-1111-111111111111
```
