---
subcategory: "Kubernetes"
page_title: "Scaleway: scaleway_k8s_acl"
---

# Resource: scaleway_k8s_acl

Creates and manages Scaleway Kubernetes Cluster authorized IPs.
For more information, please refer to the [API documentation](https://www.scaleway.com/en/developers/api/kubernetes/#path-access-control-list-add-new-acls).

~> **Important:** When creating a Cluster, it comes with a default ACL rule allowing all ranges `0.0.0.0/0`.
Defining custom ACLs with Terraform will overwrite this rule, but it will be recreated automatically when deleting the ACL resource.

## Example Usage

### Basic

```terraform
resource "scaleway_vpc_private_network" "acl_basic" {}

resource "scaleway_k8s_cluster" "acl_basic" {
	name = "acl-basic"
	version = "1.32.2"
	cni = "cilium"
	delete_additional_resources = true
	private_network_id = scaleway_vpc_private_network.acl_basic.id
}

resource "scaleway_k8s_acl" "acl_basic" {
	cluster_id = scaleway_k8s_cluster.acl_basic.id
	acl_rules {
		ip = "1.2.3.4/32"
		description = "Allow 1.2.3.4"
	}
	acl_rules {
		scaleway_ranges = true
		description = "Allow all Scaleway ranges"
	}
}
```

### Full-isolation

```terraform
resource "scaleway_vpc_private_network" "acl_basic" {}

resource "scaleway_k8s_cluster" "acl_basic" {
	name = "acl-basic"
	version = "1.32.2"
	cni = "cilium"
	delete_additional_resources = true
	private_network_id = scaleway_vpc_private_network.acl_basic.id
}

resource "scaleway_k8s_acl" "acl_basic" {
	cluster_id = scaleway_k8s_cluster.acl_basic.id
	no_ip_allowed = true
}
```

## Argument Reference

The following arguments are supported:

- `cluster_id` - (Required) UUID of the cluster. The ID of the cluster is also the ID of the ACL resource, as there can only be one per cluster.

~> **Important:** Updates to `cluster_id` will recreate the ACL.

- `no_ip_allowed` - (Optional) If set to true, no IP will be allowed and the cluster will be in full-isolation.

~> **Important:** This field cannot be set to true if the `acl_rules` block is defined.

- `acl_rules` - (Optional) A list of ACLs (structure is described below)

~> **Important:** This block cannot be defined if the `no_ip_allowed` field is set to true.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the ACL rule should be created.

The `acl_rules` block supports:

- `ip` - (Optional) The IP range to whitelist in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)

~> **Important:** If the `ip` field is set, `scaleway_ranges` cannot be set to true in the same rule.

- `scaleway_ranges` - (Optional) Allow access to cluster from all Scaleway ranges as defined in [Scaleway Network Information - IP ranges used by Scaleway](https://www.scaleway.com/en/docs/console/account/reference-content/scaleway-network-information/#ip-ranges-used-by-scaleway).
Only one rule with this field set to true can be added.

~> **Important:** If the `scaleway_ranges` field is set to true, the `ip` field cannot be set on the same rule.

- `description` - (Optional) A text describing this rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the ACL resource. It is the same as the ID of the cluster.

~> **Important:** Kubernetes ACLs' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `acl_rules.#.id` - The ID of each individual ACL rule.

## Import

Kubernetes ACLs can be imported using the `{region}/{cluster-id}`, e.g.

```bash
terraform import scaleway_k8s_acl.acl01 fr-par/11111111-1111-1111-1111-111111111111
```