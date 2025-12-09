---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_head_stage"
---

# Resource: scaleway_edge_services_head_stage

Sets the Scaleway Edge Services head stage of your pipeline.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_pipeline" "main" {
  name        = "my-edge_services-pipeline"
  description = "pipeline description"
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

```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `head_stage_id` - (Required) The ID of head stage of the pipeline.

## Attributes Reference

No additional attributes are exported.

## Import

Head stages can be imported using the `{id}`, e.g.

```bash
terraform import scaleway_edge_services_head_stage.main 11111111-1111-1111-1111-111111111111
```
