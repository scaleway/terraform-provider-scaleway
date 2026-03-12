---
subcategory: "Kubernetes"
page_title: "Scaleway: scaleway_k8s_version"
---

# scaleway_k8s_version

The [`scaleway_k8s_version`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/data-sources/k8s_version) data source is used to retrieve information about a Kubernetes version.

Refer to the Kubernetes [documentation](https://www.scaleway.com/en/docs/compute/kubernetes/) and [API documentation](https://www.scaleway.com/en/developers/api/kubernetes/) for more information.

You can also use the [scaleway-cli](https://github.com/scaleway/scaleway-cli) with `scw k8s version list` to list all available versions.



## Example Usage

```terraform
# Use the latest version
data "scaleway_k8s_version" "latest" {
  name = "latest"
}
```

```terraform
# Use a specific version
data "scaleway_k8s_version" "by_name" {
  name = "1.26.0"
}
```



## Argument Reference

- `name` - (Required) The name of the Kubernetes version.
- `region` - (Defaults to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#regions) in which the version exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the version.

~> **Important:** Kubernetes versions' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{name}`, e.g. `fr-par/1.1.1`

- `available_cnis` - The list of supported Container Network Interface (CNI) plugins for this version.
- `available_container_runtimes` - The list of supported container runtimes for this version.
- `available_feature_gates` - The list of supported feature gates for this version.
