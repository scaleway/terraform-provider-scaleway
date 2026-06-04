---
page_title: "Scaleway: scaleway_lb_backend"
subcategory: "Load Balancer"
description: |-
  Lists Scaleway Load Balancer Backends for one or more Load Balancers.
---

# Resource: scaleway_lb_backend

Lists Scaleway Load Balancer Backends for one or more Load Balancers.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/network/load-balancer/).

## Example Usage

```terraform
# List all backends for a specific Load Balancer
list "scaleway_lb_backend" "by_lb" {
  provider = scaleway

  config {
    zones  = ["fr-par-1"]
    lb_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List backends filtered by name across multiple Load Balancers
list "scaleway_lb_backend" "by_name" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
    lb_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    name = "my-backend"
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `lb_ids` - (Required) Load Balancer IDs to list backends for. Accepts either zonal IDs (e.g. `fr-par-1/uuid`) or plain UUIDs.
- `name` - (Optional) Name of the backend to filter for.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones. The cross-product of `zones` × `lb_ids` is queried.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Backend:

- `id` - The ID of the Backend.
- `name` - The name of the Backend.
- `lb_id` - The ID of the Load Balancer.
- `forward_protocol` - The forwarding protocol (`tcp` or `http`).
- `forward_port` - The port backends listen on.
- `forward_port_algorithm` - The load balancing algorithm.
- `sticky_sessions` - The sticky session type.
- `sticky_sessions_cookie_name` - The cookie name for sticky sessions.
- `server_ips` - The IP addresses of the backend servers.
- `health_check_port` - The health check port.
- `health_check_max_retries` - The maximum health check retries.
- `health_check_timeout` - The health check timeout.
- `health_check_delay` - The delay between health checks.
