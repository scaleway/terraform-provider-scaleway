---
page_title: "Using Elastic Metal servers in a Kubernetes cluster"
description: |-
How to add and/or run Scaleway baremetal servers as nodes on a multicloud Kubernetes cluster
---

# How to use Elastic Metal servers with a Kosmos multicloud cluster

In this guide you will learn how to deploy baremetal nodes on your Kubernetes cluster instead of instances. To be able
to do this, you need to have a [Kosmos multicloud cluster](../resources/k8s_cluster.md#multicloud) instead of a Kapsule
cluster, and add your [pools](../resources/k8s_pool.md) as the "external" node type.

## Prerequisites

* An SSH key, required to connect to your Elastic Metal server.
    * If you need help generating an SSH key, visit [this tutorial](https://www.scaleway.com/en/docs/console/my-project/how-to/create-ssh-key/).
    * You can view your SSH keys in the [console](https://console.scaleway.com/project/credentials) and add a new one.

## Setup

```hcl
resource "scaleway_k8s_cluster" "multicloud" {
  name    = "multicloud-cluster"
  type    = "multicloud"
  version = "1.26.0"
  cni     = "kilo"
  region  = "fr-par"
  delete_additional_resources = false
}

resource "scaleway_k8s_pool" "pool" {
  cluster_id  = scaleway_k8s_cluster.multicloud.id
  name        = "multicloud-pool"
  node_type   = "external"
  size        = 1
  region      = "fr-par"
}

resource "scaleway_baremetal_server" "server" {
  offer       = "EM-B112X-SSD"                          # The name of the Elastic Metal offer
  os          = "03b7f4ba-a6a1-4305-984e-b54fafbf1681"  # The ID of the OS, here the ID is for ubuntu_focal
  ssh_key_ids = [scaleway_iam_ssh_key.key.id]           # The list of SSH key IDs allowed to connect to the server
  zone        = "fr-par-2"
}

resource "scaleway_iam_ssh_key" "key" {
  name = "ssh-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}
```

### Notes

* `kilo` is the only CNI compatible with multicloud clusters
* Rather than giving a raw value for the `offer` and `os` fields of the baremetal server, you can use the dedicated
datasources. Ths will also allow you to check their availability in the zone you want to work with.
    * See the [baremetal offer datasource](../data-sources/baremetal_offer.md)
    * See the [baremetal os datasource](../data-sources/baremetal_os.md)
    * For more information on baremetal servers specs, visit the [resource documentation](../resources/baremetal_server.md)
* If you want to link any already existing resource, you can import it to the Terraform state by running :

   ```bash
   terraform import <scaleway_resource_type> <locality>/<id>
   ```

## Configure the Elastic Metal server

1. Get your server's public IP and SSH to the server :

    ```bash
    ssh <user>@<baremetal_server_ip>
    ```

2. Download the multicloud-init script :

    ```bash
    wget https://scwcontainermulticloud.s3.fr-par.scw.cloud/multicloud-init.sh && chmod +x multicloud-init.sh`
    ```

3. Export the required environment variables :

    ```bash
    export POOL_ID=<pool_id>  REGION=<cluster_region>  SCW_SECRET_KEY=<secret_key>`
    ```

4. Execute the script to attach the node to the multicloud pool :

    ```bash
    sudo ./multicloud-init.sh -p $POOL_ID -r $REGION -t $SCW_SECRET_KEY
    ```
