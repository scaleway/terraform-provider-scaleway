---
page_title: "Scaleway: scaleway_lb_frontends"
description: |-
Gets information about multiple Load Balancer Frontends.
---

# scaleway_lb_frontends

Gets information about multiple Load Balancer Frontends.

## Example Usage

```hcl
# Find frontends that share the same LB ID
data "scaleway_lb_frontends" "byLBID" {
  lb_id      = "${scaleway_lb.lb01.id}"
}
# Find frontends by LB ID and name
data "scaleway_lb_frontends" "byLBID_and_name" {
  lb_id      = "${scaleway_lb.lb01.id}"
  name       = "tf-backend-datasource"
}
```

## Argument Reference

- `lb_id` - (Required) The load-balancer ID this frontend is attached to. frontends with a LB ID like it are listed.

- `name` - (Optional) The frontend name used as filter. Frontends with a name like it are listed.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which frontends exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `frontends` - List of found frontends
    - `id` - The associated frontend ID.
        ~> **Important:** LB frontends' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
    - `inbound_port` - TCP port to listen on the front side.
    - `created_at` - The date at which the frontend was created (RFC 3339 format).
    - `update_at` - The date at which the frontend was last updated (RFC 3339 format).
    - `backend_id` - The load-balancer backend ID this frontend is attached to.
    - `timeout_client` - Maximum inactivity time on the client side.
    - `certificate_ids` - List of Certificate IDs that should be used by the frontend.
    - `enable_http3` - Activates HTTP/3 protocol.
