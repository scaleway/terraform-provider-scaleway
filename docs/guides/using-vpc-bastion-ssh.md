---
page_title: "Using Scaleway Bastion SSH"
description: |-
  Using Scaleway Bastion SSH config.
---

# Using Scaleway VPC Bastion SSH config

In this guide you'll learn how to deploy Scaleway Bastion with to your virtual network using the Scaleway Terraform provider.
After Bastion is deployed, you can connect (SSH) to virtual machines in the virtual network via Bastion using the private IP address of the VM.
When you connect to a VM, it doesn't need a public IP address, client software, agent, or a special configuration.

## Prerequisites

*	A Virtual Machine in a VPC PrivateNetwork. Check our example below.

1. When you deploy Bastion, the values are pulled from the Network in which your VM resides.
   1. This VM doesn't become a part of the Bastion deployment itself, but you do connect to it later in the exercise.

2. If you can don't have any VM connected on the private network please use the `scaleway_instance_private_nic` or the attribute `private_network` on `scaleway_instance_server` to connected to.

3. If your VMs are attached to any `scaleway_instance_ip`. You should detach them.

  **Note**: You should keep your VMs and Private Network on the same zone. e.g. `fr-par-1`

```hcl
# NUMBER OF VIRTUAL MACHINES
variable "machine_count" {
  description = "Number of virtual machines in private network"
  default = 3
}

# SCALEWAY VPC PRIVATE NETWORK
resource scaleway_vpc_private_network "pn" {
  name = "the-private-network"
}

# SCALEWAY VPC VIRTUAL MACHINES
resource scaleway_instance_server "servers" {
  count	= var.machine_count
  name 	= "machine${count.index}"
  image = "ubuntu_focal"
  type  = "DEV1-S"
}

# SCALEWAY INSTANCES PRIVATE NETWORK CONNECTION
resource scaleway_instance_private_nic "nic" {
  count              = length(scaleway_instance_server.servers)
  private_network_id = scaleway_vpc_private_network.pn.id
  server_id          = scaleway_instance_server.servers[count.index].id
}
```

```hcl

resource scaleway_vpc_public_gateway_ip "pgw_ip" {
}

resource scaleway_vpc_public_gateway "pgw" {
  type = "VPC-GW-S"
  bastion_enabled = true
  ip_id = scaleway_vpc_public_gateway_ip.pgw_ip.id
}
```

```hcl
resource scaleway_vpc_public_gateway_dhcp "dhcp" {
  subnet = "192.168.1.0/24"
  dns_local_name = scaleway_vpc_private_network.pn.name
}

resource scaleway_vpc_gateway_network "gn" {
  gateway_id         = scaleway_vpc_public_gateway.pgw.id
  private_network_id = scaleway_vpc_private_network.pn.id
  dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp.id
  enable_dhcp = true
}
```
