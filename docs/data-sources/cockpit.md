---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit"
---
# scaleway_cockpit


~> **Important:**  The data source `scaleway_cockpit` has been deprecated and will no longer be supported. Instead, use resource `scaleway_cockpit`.

-> **Note:**
As of April 2024, Cockpit has introduced [regionalization](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#region) to offer more flexibility and resilience.
If you have created customized dashboards with data for your Scaleway resources before April 2024, you will need to update your queries in Grafana, with the new regionalized [data sources](../resources/cockpit_source.md).

The `scaleway_cockpit` data source is used to retrieve information about a Scaleway Cockpit associated with a given Project. This can be the default Project or a specific Project identified by its ID.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Retrieve a Cockpit

The following commands allow you to:

- get information on the Cockpit associated with your Scaleway default Project
- get information on the Cockpit associated with a specific Scaleway Project

```hcl
// Get the default Project's Cockpit
data "scaleway_cockpit" "main" {}
```

```hcl
// Get a specific Project's Cockpit
data "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_cockpit` data source to filter and retrieve the desired Cockpit.

- `project_id` - Specifies the ID of the Scaleway Project that the Cockpit is associated with. If not specified, it defaults to the Project ID specified in the [provider configuration](../index.md#project_id).

- `plan` - (Optional) Specifies the name or ID of the pricing plan to use.


## Attributes Reference

The `scaleway_cockpit` data source exports certain attributes once the Cockpit information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `plan_id` - (Deprecated) ID of the current pricing plan
- `endpoints` - (Deprecated) A list of [endpoints](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#endpoints) related to Cockpit, each with specific URLs:
    - `metrics_url` - (Deprecated) URL for [metrics](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#metric) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `logs_url` - (Deprecated) URL for [logs](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#logs) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `alertmanager_url` - (Deprecated) URL for the [Alert manager](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#alert-manager).
    - `grafana_url` - (Deprecated) URL for Grafana.
