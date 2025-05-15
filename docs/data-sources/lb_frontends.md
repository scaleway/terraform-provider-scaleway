---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_frontends"
---

# scaleway_lb_frontends

Gets information about multiple Load Balancer frontends.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/reference-content/configuring-frontends/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-frontends).

## Example Usage

```hcl
# Find frontends that share the same LB ID
data "scaleway_lb_frontends" "byLBID" {
  lb_id = "${scaleway_lb.lb01.id}"
}
# Find frontends by LB ID and name
data "scaleway_lb_frontends" "byLBID_and_name" {
  lb_id = "${scaleway_lb.lb01.id}"
  name  = "tf-frontend-datasource"
}
```

## Argument Reference

- `lb_id` - (Required) The Load Balancer ID this frontend is attached to. Frontends with a matching ID are listed.

- `name` - (Optional) The frontend name to filter for. Frontends with a matching name are listed.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the frontends exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `frontends` - List of retrieved frontends
    - `id` - The ID of the associated frontend.
        ~> **Important:** LB frontend IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
    - `inbound_port` - TCP port the frontend listens to.
    - `created_at` - The date on which the frontend was created (RFC 3339 format).
    - `update_at` - The date aont which the frontend was last updated (RFC 3339 format).
    - `backend_id` - The Load Balancer backend ID this frontend is attached to.
         ~> **Important:** Load Balancer backend IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
    - `timeout_client` - Maximum inactivity time on the client side.
    - `certificate_ids` - List of certificate IDs that are used by the frontend.
    - `enable_http3` - Whether HTTP/3 protocol is activated.
