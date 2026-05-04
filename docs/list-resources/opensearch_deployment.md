---
page_title: "Scaleway: scaleway_opensearch_deployment"
subcategory: "Databases"
description: |-
  Lists Scaleway OpenSearch (SearchDB) deployments across regions and projects.
---

# Resource: scaleway_opensearch_deployment



For more information, see the [product documentation](https://www.scaleway.com/en/docs/opensearch/).


## Example Usage

```terraform
# List OpenSearch deployments across all regions and all projects
list "scaleway_opensearch_deployment" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List OpenSearch deployments filtered by name prefix
list "scaleway_opensearch_deployment" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-opensearch"
  }
```

```terraform
# List OpenSearch deployments in a specific region for a specific project
list "scaleway_opensearch_deployment" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List OpenSearch deployments filtered by tag
list "scaleway_opensearch_deployment" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
```

```terraform
# List OpenSearch deployments filtered by engine version
list "scaleway_opensearch_deployment" "by_version" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
    version     = "2.15"
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the OpenSearch deployment to filter on.
- `tags` - (Optional) Tags of the OpenSearch deployment to filter on.
- `organization_id` - (Optional) Organization ID of the OpenSearch deployment to filter on.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.
- `version` - (Optional) OpenSearch engine version to filter on (same value as the deployment `version` attribute, e.g. `"2.15"`).

## Attributes Reference

Each result corresponds to one OpenSearch deployment and exposes the same attributes as the [`scaleway_opensearch_deployment` resource](../resources/opensearch_deployment.md).
