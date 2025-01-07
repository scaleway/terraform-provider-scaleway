---
page_title: "Using Scaleway SSH Bastion"
---

# How to use pass to depecated resource cockpit to new infra <- change ici le titre pour que se soit plus claire 

-> voici l'explication : ducoup je fais un guide pour pouvoir supprimer la resource cockpit des terraforms et utiliser la nouvelle resource source, explique moi cela bien en anglais
-> **Note:**
Cockpit plans scheduled for deprecation on January 1st 2025.
The retention period previously set for your logs and metrics will remain the same after that date.
You will be able to edit the retention period for your metrics, logs, and traces for free during Beta.


## Prerequisites

d'abord il faut s'assurer d'avoir la dernier version du provider 
-> **Note:** Before upgrading to `v2+`, it is recommended to upgrade to the most recent `1.X` version of the provider (`v1.17.2`) and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected change or deprecation notice.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/language/providers/configuration#version-provider-versions).
If you are following these recommendations, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

Update to latest `1.X` version:

```hcl
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
      version = "~> 1.17"
    }
  }
}

provider "scaleway" {
  # ...
}
```

Update to latest 2.X version:

## Set up your Public Gateway

Public Gateways sit at the border of Private Networks and allow you to enable the bastion.
You can also choose your port of preference on `bastion_port` option. The default port is `61000`

You can check the types of gateways currently supported via our CLI.

```shell
scw vpc-gw gateway-type list
```

Example:

```hcl
resource scaleway_vpc_public_gateway "pgw" {
  type = "VPC-GW-S"
  bastion_enabled = true
  ip_id = scaleway_vpc_public_gateway_ip.pgw_ip.id
}
```

## Configure your DHCP on your subnet

The [DHCP](https://fr.wikipedia.org/wiki/Dynamic_Host_Configuration_Protocol) server sets the IPv4 address dynamically,
which is required to communicate over the private network.

The `dns_local_name` is the [TLD](https://en.wikipedia.org/wiki/Top-level_domain), the value by default is `priv`.
This is used to resolve your Instance on a Private Network.

In order to resolve the Instances using your Bastion you should set the `dns_local_name` with `scaleway_vpc_private_network.pn.name`.

Please check our API [documentation](https://www.scaleway.com/en/developers/api/public-gateway/#path-dhcp-create-a-dhcp-configuration) for more details.

```hcl
resource scaleway_vpc_public_gateway_dhcp "dhcp" {
  subnet = "192.168.1.0/24"
  dns_local_name = scaleway_vpc_private_network.pn.name
}
```

## Attach your VPC Gateway Network to a Private Network

To enable DHCP on this Private Network you must set `enable_dhcp` and `dhcp_id`.
Do not set the `address` attribute.

```hcl
resource scaleway_vpc_gateway_network "gn" {
  gateway_id          = scaleway_vpc_public_gateway.pgw.id
  private_network_id  = scaleway_vpc_private_network.pn.id
  dhcp_id             = scaleway_vpc_public_gateway_dhcp.dhcp.id
  enable_dhcp         = true
}
```

## Config my Bastion config

You should add your config on your local config file e.g: `~/.ssh/config`

```
Host *.myprivatenetwork
ProxyJump bastion@<your-public-ip>:<bastion_port>
```

Then try to connect to it:

```shell
ssh root@<vm-name>.myprivatenetwork
```

For further information using our console please check [our dedicated documentation](https://www.scaleway.com/en/docs/network/vpc/how-to/use-ssh-bastion/).