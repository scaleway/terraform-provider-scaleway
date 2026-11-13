---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_cache_stage"
---

# scaleway_edge_services_cache_stage (Data Source)

Gets information about an Edge Services cache stage.

A cache stage defines the caching behavior for an Edge Services pipeline, including TTL and whether cookies are included in cache keys.

## Example Usage

```terraform
# Retrieve an Edge Services cache stage by its ID
data "scaleway_edge_services_cache_stage" "by_id" {
  cache_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services cache stage by pipeline ID
data "scaleway_edge_services_cache_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
```



## Argument Reference

One of `cache_stage_id` or filter arguments must be specified.

- `cache_stage_id` - (Optional) The ID of the cache stage. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `cache_stage_id`):

- `pipeline_id` - (Required when `cache_stage_id` is not set) The ID of the pipeline.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_cache_stage` [resource](../resources/edge_services_cache_stage.md).
