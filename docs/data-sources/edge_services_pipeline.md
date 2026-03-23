---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_pipeline"
---

# scaleway_edge_services_pipeline (Data Source)

Gets information about an Edge Services pipeline.

A pipeline is the top-level resource that groups together all the stages (DNS, TLS, cache, backend, etc.) of an Edge Services configuration.

## Example Usage

```terraform
# Retrieve an Edge Services pipeline by its ID
data "scaleway_edge_services_pipeline" "by_id" {
  pipeline_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services pipeline by name
data "scaleway_edge_services_pipeline" "by_name" {
  name = "my-pipeline"
}
```



## Argument Reference

One of `pipeline_id` or filter arguments must be specified.

- `pipeline_id` - (Optional) The ID of the pipeline. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `pipeline_id`):

- `name` - (Optional) The pipeline name to filter for.
- `project_id` - (Optional) The ID of the project to filter for.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_pipeline` [resource](../resources/edge_services_pipeline.md).
