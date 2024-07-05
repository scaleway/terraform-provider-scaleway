---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_token"
---

# Resource: scaleway_cockpit_token

Creates and manages Scaleway Cockpit Tokens.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#tokens).

## Example Usage

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

- `name` - (Required) The name of the token.
- `scopes` - (Optional) Allowed scopes.
    - `query_metrics` - (Defaults to `false`) Query metrics.
    - `write_metrics` - (Defaults to `true`) Write metrics.
    - `setup_metrics_rules` - (Defaults to `false`) Setup metrics rules.
    - `query_logs` - (Defaults to `false`) Query logs.
    - `write_logs` - (Defaults to `true`) Write logs.
    - `setup_logs_rules` - (Defaults to `false`) Setup logs rules.
    - `setup_alerts` - (Defaults to `false`) Setup alerts.
    - `query_traces` - (Defaults to `false`) Query traces.
    - `write_traces` - (Defaults to `false`) Write traces.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the cockpit token.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the cockpit token.

~> **Important:** cockpit tokens' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

- `secret_key` - The secret key of the token.

## Import

Cockpits tokens can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_cockpit_token.main fr-par/11111111-1111-1111-1111-111111111111
```
