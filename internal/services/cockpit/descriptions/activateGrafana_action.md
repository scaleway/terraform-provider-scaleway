The [`scaleway_cockpit_activate_grafana`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_activate_grafana) action provisions Grafana for a Cockpit project using IAM authentication.

After Grafana users were deprecated, Grafana must be accessed at least once with an IAM secret key before actions such as [`scaleway_cockpit_grafana_sync_data_sources`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/cockpit_grafana_sync_data_sources) can succeed. This action performs that first access (`GET {grafana_url}/api/org` with `X-Auth-Token`) so you do not need a manual curl.

Refer to the Cockpit [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api/) for more information.
