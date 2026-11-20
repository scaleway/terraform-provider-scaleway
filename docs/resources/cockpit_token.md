---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_token"
---

# Resource: scaleway_cockpit_token

~> **Important:** The resource `scaleway_cockpit_token` is deprecated.
Please use IAM authentication instead. You will find an implementation example in the [IAM authentication](#authenticating-to-cockpit-with-iam) section below.

The `scaleway_cockpit_token` resource allows you to create and manage your Cockpit [tokens](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#tokens).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Authenticating to Cockpit with IAM

Prefer an IAM application and API key over `scaleway_cockpit_token`. Use the secret key in the `X-Auth-Token` header when pushing metrics/logs or querying Loki.

```terraform
data "scaleway_account_project" "default" {
  name = "default"
}

resource "scaleway_iam_application" "cockpit_auth" {
  name = "cockpit-auth"
}

resource "scaleway_iam_policy" "observability" {
  application_id = scaleway_iam_application.cockpit_auth.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["ObservabilityFullAccess"]
  }
}

resource "scaleway_iam_api_key" "api_key" {
  application_id = scaleway_iam_application.cockpit_auth.id
}

resource "scaleway_cockpit_source" "metrics" {
  name           = "my-metrics"
  type           = "metrics"
  retention_days = 31
}

resource "scaleway_cockpit_source" "logs" {
  name           = "my-logs"
  type           = "logs"
  retention_days = 7
}

output "secret_key" {
  value     = scaleway_iam_api_key.api_key.secret_key
  sensitive = true
}

output "metrics_push_url" {
  value = scaleway_cockpit_source.metrics.push_url
}

output "logs_url" {
  value = scaleway_cockpit_source.logs.url
}

output "logs_push_url" {
  value = scaleway_cockpit_source.logs.push_url
}
```

Use the IAM secret key in the `X-Auth-Token` header (not `Authorization: Bearer`) when calling Cockpit endpoints — for example querying Loki, or configuring Prometheus remote write / Loki push agents against `push_url`:

```shell
curl -H "X-Auth-Token: $(terraform output -raw secret_key)" \
  "$(terraform output -raw logs_url)/loki/api/v1/labels"
```

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
- `region` - (Optional, Computed, Defaults to the region specified in the [provider configuration](../index.md#arguments-reference)) The [region](../guides/regions_and_zones.md#regions) where the Cockpit token is located.
- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#arguments-reference)) The ID of the Project the Cockpit is associated with.

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
