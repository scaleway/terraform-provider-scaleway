---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_waf_stage"
---

# scaleway_edge_services_waf_stage (Data Source)

Gets information about an Edge Services WAF (Web Application Firewall) stage.

A WAF stage provides web application firewall protection for an Edge Services pipeline, inspecting HTTP requests and blocking malicious traffic based on a configurable paranoia level.

## Example Usage

```terraform
# Retrieve an Edge Services WAF stage by its ID
data "scaleway_edge_services_waf_stage" "by_id" {
  waf_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services WAF stage by pipeline ID
data "scaleway_edge_services_waf_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
```



## Argument Reference

One of `waf_stage_id` or filter arguments must be specified.

- `waf_stage_id` - (Optional) The ID of the WAF stage. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `waf_stage_id`):

- `pipeline_id` - (Required when `waf_stage_id` is not set) The ID of the pipeline.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_waf_stage` [resource](../resources/edge_services_waf_stage.md).
