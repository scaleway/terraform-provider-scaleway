---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_routes"
---

# scaleway_lb_routes

Gets information about multiple Load Balancer routes.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/how-to/create-manage-routes/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-route).

## Example Usage

```hcl
# Find routes that share the same frontend ID
data "scaleway_lb_routes" "by_frontendID" {
  frontend_id = scaleway_lb_frontend.frt01.id
}
# Find routes by frontend ID and zone
data "scaleway_lb_routes" "my_key" {
  frontend_id = "11111111-1111-1111-1111-111111111111"
  zone        = "fr-par-2"
}
```

## Argument Reference

- `frontend_id` - (Optional) The frontend ID (the origin of the redirection), to filter for. Routes with a matching frontend ID are listed.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the routes exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `routes` - List of retrieved routes
    - `id` - The associated route ID.
    - `backend_id` - The backend ID to redirect to
    - `created_at` - The date on which the route was created (RFC 3339 format).
    - `update_at` - The date on which the route was last updated (RFC 3339 format).
    - `match_sni` - Server Name Indication TLS extension field from an incoming connection made via an SSL/TLS transport layer.
    - `match_host_header` - Specifies the host of the server to which the request is being sent.