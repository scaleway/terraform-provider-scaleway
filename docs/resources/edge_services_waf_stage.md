---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_waf_stage"
---

# Resource: scaleway_edge_services_waf_stage

Creates and manages Scaleway Edge Services WAF Stages.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_waf_stage" "main" {
  pipeline_id    = scaleway_edge_services_pipeline.main.id
  mode           = "enable"
  paranoia_level = 3
}
```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `paranoia_level` - (Required) The sensitivity level (`1`,`2`,`3`,`4`) to use when classifying requests as malicious. With a high level, requests are more likely to be classed as malicious, and false positives are expected. With a lower level, requests are more likely to be classed as benign.
- `backend_stage_id` - (Optional) The ID of the backend stage to forward requests to after the WAF stage.
- `mode` - (Optional) The mode defining WAF behavior (`disable`/`log_only`/`enable`).
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the WAF stage is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the WAF stage (UUID format).
- `created_at` - The date and time of the creation of the WAF stage.
- `updated_at` - The date and time of the last update of the WAF stage.

## Import

WAF stages can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_edge_services_waf_stage.basic 11111111-1111-1111-1111-111111111111
```
