---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_grafana_user"
---

# Resource: scaleway_cockpit_grafana_user

~> **Deprecated:** This resource is deprecated and will be removed on **January 1st, 2026**.

~> **Migration Guide:** Grafana authentication is now managed through [Scaleway IAM (Identity and Access Management)](https://www.scaleway.com/en/docs/identity-and-access-management/iam/). To access your Grafana instance, use the [`scaleway_cockpit_grafana` data source](../data-sources/cockpit_grafana.md) to retrieve the Grafana URL and authenticate using your Scaleway IAM credentials.

The `scaleway_cockpit_grafana_user` resource allows you to create and manage [Grafana users](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#grafana-users) in Scaleway Cockpit.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Migration to IAM Authentication

Instead of managing Grafana users, retrieve your Grafana URL using the data source:

```terraform
# Old approach (deprecated)
# resource "scaleway_cockpit_grafana_user" "main" {
#   project_id = scaleway_account_project.project.id
#   login      = "my-awesome-user"
#   role       = "editor"
# }

# New approach - Use IAM authentication
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

output "grafana_url" {
  value       = data.scaleway_cockpit_grafana.main.grafana_url
  description = "Access Grafana using your Scaleway IAM credentials"
}
```

### Programmatic access with the Grafana provider

To automate Grafana configuration (e.g., dashboards, alerting) with Terraform, reuse your Scaleway IAM secret as an `X-Auth-Token` header. The Grafana provider must run in `anonymous` mode because user/password authentication is deprecated.

```terraform
variable "scaleway_secret_key" {
  description = "Scaleway IAM secret key used for both the Scaleway and Grafana providers"
  type        = string
  sensitive   = true
}

data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

provider "grafana" {
  url  = data.scaleway_cockpit_grafana.main.grafana_url
  auth = "anonymous"

  http_headers = {
    "X-Auth-Token" = var.scaleway_secret_key
  }
}
```

The header `X-Auth-Token` is mandatory when Grafana users are disabled. Store the IAM secret key securely (environment variable, secrets manager, etc.) and avoid committing it to version control.

### Create a Grafana user (Deprecated)

The following command allows you to create a Grafana user within a specific Scaleway Project.

```terraform
resource "scaleway_account_project" "project" {
  name = "test project grafana user"
}

resource "scaleway_cockpit_grafana_user" "main" {
  project_id = scaleway_account_project.project.id
  login      = "my-awesome-user"
  role       = "editor"
}
```

## Argument Reference

This section lists the arguments that are supported:

- `login` - (Required) The username of the Grafana user. The `admin` user is not yet available for creation. You need your Grafana username to log in to Grafana and access your dashboards.
- `role` - (Required) The role assigned to the Grafana user. Must be `editor` or `viewer`.
- `project_id` - (Defaults to Project ID specified in the [provider configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `password` - The password of the Grafana user.
- `grafana_url` - URL for Grafana.

## Import

This section explains how to import Grafana users using the ID of the Project associated with Cockpit, and the Grafana user ID in the `{project_id}/{grafana_user_id}` format.

```bash
terraform import scaleway_cockpit_grafana_user.main 11111111-1111-1111-1111-111111111111/2
```
