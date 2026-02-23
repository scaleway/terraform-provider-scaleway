---
subcategory: "Kafka"
page_title: "Scaleway: scaleway_kafka_cluster"
---

# scaleway_kafka_cluster

The [`scaleway_kafka_cluster`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/data-sources/kafka_cluster) data source is used to retrieve information about a Kafka cluster.

For more information refer to the [product documentation](https://www.scaleway.com/en/docs/managed-services/kafka/).

~> **Important:** The Kafka product is currently in Public Beta.

## Example Usage

```terraform
# Get info by cluster name
data "scaleway_kafka_cluster" "by_name" {
  name = "my-kafka-cluster"
}
```

```terraform
# Get info by cluster ID
data "scaleway_kafka_cluster" "by_id" {
  cluster_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the Kafka cluster. Only one of `name` and `cluster_id` should be specified.

- `cluster_id` - (Optional) The ID of the Kafka cluster. Only one of `name` and `cluster_id` should be specified.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the cluster exists.

- `project_id` - (Optional) The ID of the project the cluster is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Kafka cluster.

~> **Important:** Kafka cluster IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `name` - Name of the Kafka cluster.
- `version` - Kafka version (e.g., "3.9.0").
- `node_amount` - Number of nodes in the cluster.
- `node_type` - Node type used for the cluster.
- `volume_type` - Type of volume where data is stored.
- `volume_size_in_gb` - Volume size in GB.
- `status` - The status of the cluster (e.g., "ready", "creating", "configuring").
- `created_at` - Date and time of cluster creation (RFC 3339 format).
- `updated_at` - Date and time of cluster last update (RFC 3339 format).
- `tags` - List of tags associated with the cluster.

### Public Network (Computed)

~> **Note:** Public endpoints are not yet supported and this block will be empty until the feature is available.

- `public_network` - Public endpoint information.
    - `id` - The ID of the public endpoint.
    - `dns_records` - List of DNS records for the public endpoint.
    - `port` - TCP port number.

### Private Network (Computed)

When the cluster has a private network endpoint configured:

- `private_network` - Private network endpoint information.
    - `pn_id` - The private network ID.
    - `id` - The ID of the private endpoint.
    - `dns_records` - List of DNS records for the private endpoint.
    - `port` - TCP port number.
