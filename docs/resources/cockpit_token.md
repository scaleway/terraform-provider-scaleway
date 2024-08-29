---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_token"
---

# Resource: scaleway_cockpit_token

The `scaleway_cockpit_token` resource allows you to create and manage your Cockpit [tokens](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#tokens).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Use a Cockpit token

The following commands allow you to:

- create a Scaleway Project named `my-project`
- create a Cockpit token named `my-awesome-token` inside the Project
- assign `read` permissions to the token for metrics and logs
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

## Argument Reference

This section lists the arguments that are supported:

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
- `region` - (Defaults to the region specified in the [provider configuration](../index.md#region)) The [region](../guides/regions_and_zones.md#regions) where the Cockpit token is located.
- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Cockpit token.

~> **Important:** Cockpit tokens' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means that they include the region, in the `{region}/{id}` format. For example, if your token is located in the `fr-par` region, its ID would look like the following: `fr-par/11111111-1111-1111-1111-111111111111`.

- `secret_key` - The secret key of the token.

## Import

This section explains how to import a Cockpit token using the `{region}/{id}` format.

```bash
terraform import scaleway_cockpit_token.main fr-par/11111111-1111-1111-1111-111111111111
```
