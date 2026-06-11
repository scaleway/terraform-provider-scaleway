---
subcategory: "OpenSearch"
page_title: "Scaleway: scaleway_opensearch_deployment"
---

# scaleway_opensearch_deployment

Gets information about an OpenSearch deployment.

For more information refer to the [product documentation](https://www.scaleway.com/en/docs/opensearch/).

## Example Usage

```hcl
# Get info by name
data "scaleway_opensearch_deployment" "by_name" {
  name = "my-opensearch-cluster"
}

# Get info by deployment ID
data "scaleway_opensearch_deployment" "by_id" {
  deployment_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the OpenSearch deployment. Only one of `name` and `deployment_id` should be specified.
- `deployment_id` - (Optional) The ID of the OpenSearch deployment. Only one of `name` and `deployment_id` should be specified.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment exists.
- `project_id` - (Optional) The ID of the project the OpenSearch deployment is associated with.

## Attributes Reference

Exported attributes are the ones from `scaleway_opensearch_deployment` [resource](../resources/opensearch_deployment.md).
