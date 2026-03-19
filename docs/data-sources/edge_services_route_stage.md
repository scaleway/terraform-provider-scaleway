---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_route_stage"
---

# scaleway_edge_services_route_stage (Data Source)

Gets information about an Edge Services route stage.

A route stage defines HTTP request routing rules that forward requests to different backend stages based on method and path matching.

## Example Usage

```terraform
# Retrieve an Edge Services route stage by its ID
data "scaleway_edge_services_route_stage" "by_id" {
  route_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services route stage by pipeline ID
data "scaleway_edge_services_route_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
```



## Argument Reference

One of `route_stage_id` or filter arguments must be specified.

- `route_stage_id` - (Optional) The ID of the route stage. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `route_stage_id`):

- `pipeline_id` - (Required when `route_stage_id` is not set) The ID of the pipeline.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_route_stage` [resource](../resources/edge_services_route_stage.md).
