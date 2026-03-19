---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_backend_stage"
---

# scaleway_edge_services_backend_stage (Data Source)

Gets information about an Edge Services backend stage.

A backend stage defines the origin (Scaleway Object Storage bucket or Load Balancer) that Edge Services forwards requests to.

## Example Usage

```terraform
# Retrieve an Edge Services backend stage by its ID
data "scaleway_edge_services_backend_stage" "by_id" {
  backend_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services backend stage by pipeline ID
data "scaleway_edge_services_backend_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
```



## Argument Reference

One of `backend_stage_id` or filter arguments must be specified.

- `backend_stage_id` - (Optional) The ID of the backend stage.

The following filter arguments are supported (cannot be used with `backend_stage_id`):

- `pipeline_id` - (Required when `backend_stage_id` is not set) The ID of the pipeline.
- `bucket_name` - (Optional) Filter by S3 bucket name.
- `bucket_region` - (Optional) Filter by S3 bucket region.
- `lb_id` - (Optional) Filter by Load Balancer ID.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_backend_stage` [resource](../resources/edge_services_backend_stage.md).
