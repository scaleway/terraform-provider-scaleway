---
layout: "scaleway"
page_title: "Scaleway: scaleway_lb_frontend_beta"
description: |-
  Manages Scaleway Load-Balancer Frontends.
---

# scaleway_lb_frontend_beta

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancer Frontends. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples
    
### Basic

```hcl
resource "scaleway_lb_frontend_beta" "backend01" {
    lb_id = scaleway_lb_beta.lb01.id
    backend_id = scaleway_lb_backend_beta.bkd01.id
    name = "frontend01"
    inbound_port = "80"
}
```

## Arguments Reference

The following arguments are supported:

- `lb_id`                       - (Required) The load-balancer ID this backend is attached to.
- `backend_id`                  - (Required) The load-balancer backend ID this frontend is attached to.
~> **Important:** Updates to `lb_id` or `backend_id` will recreate the backend.
- `inbound_port`                - (Required) TCP port to listen on the front side.
- `name`                        - (Optional) The name of the load-balancer frontend.
- `timeout_client`              - (Optional) maximum inactivity time on the client side. (e.g.: `1s`)
- `certificate_id`              - (Required) Certificate ID that should be used by the frontend.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the loadbalancer frontend.


## Import

Load-Balancer frontend can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_frontend_beta.backend01 fr-par/11111111-1111-1111-1111-111111111111
```
