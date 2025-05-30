---
subcategory: "Elastic Metal"
page_title: "Scaleway: scaleway_flexible_ip"
---

# scaleway_flexible_ip

Gets information about a Flexible IP.

## Example Usage

```hcl
# Get info by IP address
data "scaleway_flexible_ip" "with_ip" {
  ip_address = "1.2.3.4"
}

# Get info by IP ID
data "scaleway_flexible_ip" "with_id" {
  flexible_ip_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `ip_address` - (Optional) The IP address.
  Only one of `ip_address` and `flexible_ip_id` should be specified.

- `flexible_ip_id` - (Optional) The IP ID.
  Only one of `ip_address` and `flexible_ip_id` should be specified.

- `project_id` - (Optional. Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the Flexible IP is associated with.

## Attributes Reference

Exported attributes are the ones from `flexible_ip` [resource](../resources/flexible_ip.md)
