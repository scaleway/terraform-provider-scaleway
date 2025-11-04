---
subcategory: "Audit Trail"
page_title: "Scaleway: scaleway_audit_trail_event"
---

# scaleway_audit_trail_event

Use this data source to get a list of existing Audit Trail events.
For more information refer to the [Audit Trail API documentation](https://www.scaleway.com/en/developers/api/audit-trail/).

## Example Usage

```hcl
# Retrieve all audit trail events on the default organization
data "scaleway_audit_trail_event" "find_all" {
  region = "fr-par"
}

# Retrieve audit trail events on a specific organization
data "scaleway_audit_trail_event" "find_by_org" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}

# Retrieve audit trail events on a specific project
data "scaleway_audit_trail_event" "find_by_project" {
  region = "fr-par"
  project_id = "11111111-1111-1111-1111-111111111111"
}

# Retrieve audit trail events for a specific type of resource
data "scaleway_audit_trail_event" "find_by_resource_type" {
  resource_type = "instance_server"
}

# Retrieve audit trail for a specific resource
data "scaleway_audit_trail_event" "find_by_resource_id" {
  resource_id = "11111111-1111-1111-1111-111111111111"
}

# Retrieve audit trail for a specific Scaleway product
data "scaleway_audit_trail_event" "find_by_product_name" {
  region = "nl-ams"
  product_name = "secret-manager"
}
```

## Argument Reference

- `region` - (Optional) The [region](../guides/regions_and_zones.md#regions) you want to target. Defaults to the region specified in the [provider configuration](../index.md#region).
- `organization_id` - (Optional. Defaults to [provider](../index.md#organization_id) `organization_id`) ID of the Organization containing the Audit Trail events.
- `project_id` - (Optional) ID of the Project containing the Audit Trail events.
- `resource_type` - (Optional) Type of the scaleway resources associated with the listed events. Possible values are: `secm_secret`, `secm_secret_version`, `kube_cluster`, `kube_pool`, `kube_node`, `kube_acl`, `keym_key`, `iam_user`, `iam_application`, `iam_group`, `iam_policy`, `iam_api_key`, `iam_ssh_key`, `iam_rule`, `iam_saml`, `iam_saml_certificate`, `secret_manager_secret`, `secret_manager_version`, `key_manager_key`, `account_user`, `account_organization`, `account_project`, `instance_server`, `instance_placement_group`, `instance_security_group`, `instance_volume`, `instance_snapshot`, `instance_image`, `apple_silicon_server`, `baremetal_server`, `baremetal_setting`, `ipam_ip`, `sbs_volume`, `sbs_snapshot`, `load_balancer_lb`, `load_balancer_ip`, `load_balancer_frontend`, `load_balancer_backend`, `load_balancer_route`, `load_balancer_acl`, `load_balancer_certificate`, `sfs_filesystem`, or `vpc_private_network`.
- `resource_id` - (Optional) ID of the Scaleway resource associated with the listed events.
- `product_name` - (Optional) Name of the Scaleway product in a hyphenated format.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `events` - List of Audit Trail events matching the requested criteria.
    - `id` - ID of the event. (UUID format)
    - `recorded_at` - Timestamp of the event. (RFC 3339 format)
    - `locality` - Locality of the resource attached to the event.
    - `principal_id` - ID of the user or IAM application at the origin of the event.
    - `organization_id` - ID of the Organization containing the Audit Trail events. (UUID format)
    - `project_id` - Project of the resource attached to the event. (UUID format)
    - `source_ip` - IP address at the origin of the event. (IP address)
    - `user_agent` - User Agent at the origin of the event.
    - `product_name` - Scaleway product associated with the listed events in a hyphenated format.
    - `service_name` - API name called to trigger the event.
    - `method_name` - API method called to trigger the event.
    - `resources` - List of resources attached to the event.
        - `id` - ID of the resource attached to the event. (UUID format)
        - `type` - Type of the Scaleway resource.
        - `name` - Name of the Scaleway resource.
    - `request_id` - Unique identifier of the request at the origin of the event. (UUID format)
    - `request_body` - Request at the origin of the event.
    - `status_code` - HTTP status code resulting of the API call.

