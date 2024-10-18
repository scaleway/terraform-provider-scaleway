---
subcategory: "MongoDB"
page_title: "Scaleway: scaleway_mongodb_instance"
---

# scaleway_mongodb_instance

Gets information about a MongoDB Instance.

For further information refer to the Managed Databases for MongoDB [API documentation](https://developers.scaleway.com/en/products/mongodb/api/)

## Example Usage

```hcl
# Get info by name
data "scaleway_mongodb_instance" "my_instance" {
  name = "foobar"
}

# Get info by instance ID
data "scaleway_mongodb_instance" "my_instance" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}

# Get other attributes
output "mongodb_version" {
  description = "Version of the MongoDB instance"
  value = data.scaleway_mongodb_instance.my_instance.version
}
```

## Argument Reference

- `name` - (Optional) The name of the MongoDB instance.

- `instance_id` - (Optional) The MongoDB instance ID.

  -> **Note** You must specify at least one: `name` or `instance_id`.

- `project_id` - (Optional) The ID of the project the MongoDB instance is in. Can be used to filter instances when using `name`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the MongoDB Instance exists.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the MongoDB instance is in.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the MongoDB Instance.
- `name` - The name of the MongoDB instance.
- `version` - The version of MongoDB running on the instance.
- `node_type` - The type of MongoDB node.
- `node_number` - The number of nodes in the MongoDB cluster.
- `created_at` - The date and time the MongoDB instance was created.
- `project_id` - The ID of the project the instance belongs to.
- `tags` - A list of tags attached to the MongoDB instance.
- `volume_type` - The type of volume attached to the MongoDB instance.
- `volume_size_in_gb` - The size of the attached volume, in GB.
- `public_network` - The details of the public network configuration, if applicable.

## Import

MongoDB™ instance can be imported using the `id`, e.g.

```bash
terraform import scaleway_mongodb_instance.main fr-par-1/11111111-1111-1111-1111-111111111111
```
