---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_instance_apply_maintenance"
---

# scaleway_rdb_instance_apply_maintenance (Action)

The [`scaleway_rdb_instance_apply_maintenance`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/rdb_instance_apply_maintenance) action applies pending maintenance tasks on a Managed Database instance.

Refer to the RDB [documentation](https://www.scaleway.com/en/docs/managed-databases-for-postgresql-and-mysql/) and [API documentation](https://www.scaleway.com/en/developers/api/managed-databases-for-postgresql-and-mysql/) for more information.

## Schema

### Required

- `instance_id` (String) RDB instance ID to apply maintenance on. Can be a plain UUID or a regional ID.

### Optional

- `region` (String) The region you want to attach the resource to
- `wait` (Boolean) Wait for the instance maintenance to complete before returning.
