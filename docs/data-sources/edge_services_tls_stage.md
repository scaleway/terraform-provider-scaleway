---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_tls_stage"
---

# scaleway_edge_services_tls_stage (Data Source)

Gets information about an Edge Services TLS stage.

A TLS stage manages TLS/SSL certificates for an Edge Services pipeline, supporting both managed Let's Encrypt certificates and custom certificates stored in Scaleway Secret Manager.

## Example Usage

```terraform
# Retrieve an Edge Services TLS stage by its ID
data "scaleway_edge_services_tls_stage" "by_id" {
  tls_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services TLS stage by pipeline ID
data "scaleway_edge_services_tls_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
```



## Argument Reference

One of `tls_stage_id` or filter arguments must be specified.

- `tls_stage_id` - (Optional) The ID of the TLS stage. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `tls_stage_id`):

- `pipeline_id` - (Required when `tls_stage_id` is not set) The ID of the pipeline.
- `secret_id` - (Optional) Secret ID to filter for.
- `secret_region` - (Optional) Secret region to filter for.

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_tls_stage` [resource](../resources/edge_services_tls_stage.md).
