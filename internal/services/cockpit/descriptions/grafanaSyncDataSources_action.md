The [`scaleway_cockpit_grafana_sync_data_sources`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_grafana_sync_data_sources) action triggers the synchronization of all Cockpit data sources and the alert manager across regions into Grafana.

Before syncing, the action also performs an IAM first access (`GET {grafana_url}/api/org`) so Grafana is provisioned for the project.

Refer to the Cockpit [documentation](https://www.scaleway.com/en/docs/managed-services/cockpit/) and [API documentation](https://www.scaleway.com/en/docs/observability/cockpit/api-cli/) for more information.
