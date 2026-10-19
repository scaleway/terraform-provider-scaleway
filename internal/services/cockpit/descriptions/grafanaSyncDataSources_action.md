The [`scaleway_cockpit_grafana_sync_data_sources`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_grafana_sync_data_sources) action triggers the synchronization of all Cockpit data sources and the alert manager across regions into Grafana.

Before syncing, the action performs an IAM first access (`GET {grafana_url}/api/org` with the provider secret key) so Grafana is provisioned for the project. No manual curl or console visit is required.

Refer to the Cockpit [documentation](https://www.scaleway.com/en/docs/managed-services/cockpit/) and [API documentation](https://www.scaleway.com/en/docs/observability/cockpit/api-cli/) for more information.
