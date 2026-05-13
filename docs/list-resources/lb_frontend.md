---
page_title: "Scaleway: scaleway_lb_frontend"
subcategory: "Load Balancer"
description: |-
  Lists Scaleway Load Balancer Frontends for one or more Load Balancers.
---

# Resource: scaleway_lb_frontend

Lists Scaleway Load Balancer Frontends for one or more Load Balancers.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/network/load-balancer/).

## Example Usage

```terraform
# List all frontends for a specific Load Balancer
list "scaleway_lb_frontend" "by_lb" {
  provider = scaleway

  config {
    zones  = ["fr-par-1"]
    lb_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List frontends filtered by name
list "scaleway_lb_frontend" "by_name" {
  provider = scaleway

  config {
    zones  = ["fr-par-1"]
    lb_ids = ["11111111-1111-1111-1111-111111111111"]
    name   = "my-frontend"
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `lb_ids` - (Required) Load Balancer IDs to list frontends for. Accepts either zonal IDs (e.g. `fr-par-1/uuid`) or plain UUIDs.
- `name` - (Optional) Name of the frontend to filter for.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones. The cross-product of `zones` × `lb_ids` is queried.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Frontend:

- `id` - The ID of the Frontend.
- `name` - The name of the Frontend.
- `lb_id` - The ID of the Load Balancer.
- `backend_id` - The ID of the Backend.
- `inbound_port` - The TCP port the frontend listens on.
- `timeout_client` - The maximum client inactivity time.
- `enable_http3` - Whether HTTP/3 is enabled.
- `certificate_id` - The ID of the TLS certificate.
- `certificate_ids` - The IDs of all TLS certificates.
- `connection_rate_limit` - The maximum connection rate (connections/s).
- `enable_access_logs` - Whether access logs are enabled.
- `acl` - The ACL rules attached to the frontend.
- `created_at` - The date and time of creation.
- `updated_at` - The date and time of the last update.
