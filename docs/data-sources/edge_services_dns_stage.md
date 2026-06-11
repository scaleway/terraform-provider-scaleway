---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_dns_stage"
---

# scaleway_edge_services_dns_stage (Data Source)

Gets information about an Edge Services DNS stage.

A DNS stage defines the Fully Qualified Domain Names (FQDNs) attached to an Edge Services pipeline and links them to the next processing stage.

## Example Usage

```terraform
# Retrieve an Edge Services DNS stage by its ID
data "scaleway_edge_services_dns_stage" "by_id" {
  dns_stage_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve an Edge Services DNS stage by pipeline ID and FQDN
data "scaleway_edge_services_dns_stage" "by_fqdn" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  fqdn        = "cdn.example.com"
}
```



## Argument Reference

One of `dns_stage_id` or filter arguments must be specified.

- `dns_stage_id` - (Optional) The ID of the DNS stage. Conflicts with all filter arguments below.

The following filter arguments are supported (cannot be used with `dns_stage_id`):

- `pipeline_id` - (Required when `dns_stage_id` is not set) The ID of the pipeline.
- `fqdn` - (Optional) FQDN to filter for (in the format subdomain.example.com).

## Attributes Reference

Exported attributes are the ones from `scaleway_edge_services_dns_stage` [resource](../resources/edge_services_dns_stage.md).
