The [`scaleway_cockpit_grafana_sync_data_sources`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_grafana_sync_data_sources) action triggers the synchronization of all Cockpit data sources and the alert manager across regions into Grafana.

Grafana must be provisioned for the project before running this action. Prefer [`scaleway_cockpit_activate_grafana`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_activate_grafana) (IAM first access).

Refer to the Cockpit [documentation](https://www.scaleway.com/en/docs/managed-services/cockpit/) and [API documentation](https://www.scaleway.com/en/docs/observability/cockpit/api-cli/) for more information.
