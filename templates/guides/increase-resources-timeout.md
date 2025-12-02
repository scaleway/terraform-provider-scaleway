---
page_title: "Increasing Timeout in terraform apply"
---

# Increasing Timeout in terraform apply

When deploying resources with the Scaleway Terraform provider, operations like creating or updating instances, buckets, or other infrastructure may take longer than the default timeout durations. To avoid failures caused by these timeouts, you can configure custom timeouts per resource.

## How to set custom timeouts

Most Scaleway Terraform resources support a timeouts block where you can specify how long Terraform should wait for each operation.

### Timeout keywords supported

- `create` — Timeout duration for creating the resource.
- `update` — Timeout duration for updating the resource.
- `delete` — Timeout duration for deleting the resource.
- `default` — (Scaleway-specific) A unified timeout applied to all operations (create, update, delete) if specific keys are not set.

### Example: Using Specific Operation Timeouts

```terraform
resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_k8s_cluster" "cluster" {
  name                         = "tf-cluster"
  version                      = "1.32.3"
  cni                          = "cilium"
  private_network_id           = scaleway_vpc_private_network.pn.id
  delete_additional_resources  = false

  timeouts {
    delete = "15m"
    create = "20m"
    update = "15m"
  }
}

resource "scaleway_k8s_pool" "pool" {
  cluster_id = scaleway_k8s_cluster.cluster.id
  name       = "tf-pool"
  node_type  = "DEV1-M"
  size       = 1
}

```

### Example: Using the default Timeout

```terraform
resource "scaleway_object_bucket" "test" {
  name = "this-is-a-test"
  tags = {
    TestName = "TestAccSCW_WebsiteConfig_basic"
  }
  timeouts {
    default = "5m"
  }
}
```

If both default and one of create, update, or delete are set, the specific key overrides the default value.

If no timeouts block is set, Terraform uses the provider's internal defaults.

Not all Scaleway resources support timeouts.

Custom timeouts are useful for long-running operations but do not affect retry intervals or polling frequency.

## Official Documentation

For more details on timeout support, refer to the official [Terraform documentation](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/retries-and-customizable-timeouts).
