---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_token"
---

# Resource: scaleway_cockpit_token

The `scaleway_cockpit_token` resource allows you to create and manage your Cockpit [tokens](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#tokens).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

The following commands show you how to:

- create a Scaleway Project named `my-project`
- create a Cockpit token named `my-awesome-token` in the Project
- assign `read` permissions for metrics and logs to the token
- disable `write` permissions for metrics and logs

```terraform
resource "scaleway_account_project" "project" {
  name = "my-project"
}

resource "scaleway_cockpit_token" "main" {
  project_id = scaleway_account_project.project.id
  name       = "my-awesome-token"
}
```

```terraform
resource "scaleway_account_project" "project" {
  name = "my-project"
}

// Create a token that can read metrics and logs but not write
resource "scaleway_cockpit_token" "main" {
  project_id = scaleway_account_project.project.id
  
  name = "my-awesome-token"
  scopes {
    query_metrics = true
    write_metrics = false

    query_logs = true
    write_logs = false
  }
}
```

## Arguments reference

- `name` - (Required) The name of the token.
- `scopes` - (Optional) Scopes allowed, each with default values:
    - `query_metrics` - (Defaults to `false`) Permission to query metrics.
    - `write_metrics` - (Defaults to `true`) Permission to write metrics.
    - `setup_metrics_rules` - (Defaults to `false`) Permission to set up metrics rules.
    - `query_logs` - (Defaults to `false`) Permission to query logs.
    - `write_logs` - (Defaults to `true`) Permission to write logs.
    - `setup_logs_rules` - (Defaults to `false`) Permission to set up logs rules.
    - `setup_alerts` - (Defaults to `false`) Permission to set up alerts.
    - `query_traces` - (Defaults to `false`) Permission to query traces.
    - `write_traces` - (Defaults to `false`) Permission to write traces.
- `region` - (Defaults to the region specified in the [provider's configuration](../index.md#region)) The [region](../guides/regions_and_zones.md#regions) where the Cockpit token is located.
- `project_id` - (Defaults to the Project ID specified in the [provider's configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.

## Attributes reference

This section lists the attributes that are automatically exported when the `cockpit_token` resource is created:

- `id` - The ID of the Cockpit token.

~> **Important:** Cockpit tokens' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they include the region, in the format `{region}/{id}`. For example, if your token is located in the `fr-par` region, its ID would look like the following: `fr-par/11111111-1111-1111-1111-111111111111`.

- `secret_key` - The secret key of the token.

## Import

This section explains how to import Cockpit tokens using the `{region}/{id}` format.

```bash
terraform import scaleway_cockpit_token.main fr-par/11111111-1111-1111-1111-111111111111
```
