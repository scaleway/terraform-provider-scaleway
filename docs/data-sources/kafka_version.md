---
subcategory: "Kafka"
page_title: "Scaleway: scaleway_kafka_version"
---

# scaleway_kafka_version

The `scaleway_kafka_version` data source is used to retrieve information about an available Kafka version.

Refer to the [Kafka documentation](https://www.scaleway.com/en/docs/managed-databases/kafka/) and [API documentation](https://www.scaleway.com/en/developers/api/kafka/) for more information.

## Example Usage

```terraform
# Use the latest version
data "scaleway_kafka_version" "latest" {
  name = "latest"
}
```

```terraform
# Use a specific version
data "scaleway_kafka_version" "by_name" {
  name = "4.0.0"
}
```

## Argument Reference

- `name` - (Required) The Kafka version name. Use `latest` to retrieve the most recent available version.
- `region` - (Defaults to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#regions) in which the version exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the version, in the `{region}/{version}` format.
- `end_of_life_at` - The end-of-life date of the version (RFC 3339 format).
- `available_settings` - The cluster configuration settings available for clusters running this version. Each setting has the following attributes:
    - `name` - The setting name.
    - `hot_configurable` - Whether the setting can be applied without a restart.
    - `description` - The setting description.
    - `bool_property` - Boolean property, if the setting is a boolean. Contains `default_value`.
    - `string_property` - String property, if the setting is a string. Contains `default_value` and `string_constraint`.
    - `int_property` - Integer property, if the setting is an integer. Contains `min`, `max`, `default_value` and `unit`.
    - `float_property` - Float property, if the setting is a float. Contains `min`, `max`, `default_value` and `unit`.
