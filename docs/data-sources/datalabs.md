---
subcategory: "Datalab"
page_title: "Scaleway: scaleway_datalabs"
---

# scaleway_datalabs (Data Source)

Lists Scaleway Datalab instances.

## Example Usage

### List All

```terraform
data "scaleway_datalabs" "all" {
  region = "fr-par"
}
```

### Filtered

```terraform
data "scaleway_datalabs" "filtered" {
  region = "fr-par"
  name   = "my-datalab"
  tags   = ["production"]
}
```

## Argument Reference

- `project_id` - (Optional) The project ID to filter Datalabs by.
- `organization_id` - (Optional) The organization ID to filter Datalabs by.
- `region` - (Optional) The region to list Datalabs from.
- `name` - (Optional) The name to filter Datalabs by.
- `tags` - (Optional) The tags to filter Datalabs by.

## Attributes Reference

- `datalabs` - The list of Datalab instances.
    - `id` - The unique identifier of the Datalab instance.
    - `name` - The name of the Datalab instance.
    - `description` - The description of the Datalab instance.
    - `status` - The current status of the Datalab instance.
    - `tags` - Tags associated with the Datalab instance.
    - `region` - The region of the Datalab instance.
    - `project_id` - The project ID of the Datalab instance.
    - `spark_version` - The Spark version of the Datalab instance.
    - `has_notebook` - Whether a JupyterLab notebook is associated with the Datalab.
    - `created_at` - The creation timestamp of the Datalab instance.
    - `updated_at` - The last update timestamp of the Datalab instance.
