---
page_title: "Migrating from Legacy VPC Gateway to v2"
---

# Migrating from Legacy VPC Gateway to v2

This guide explains how to migrate from the legacy VPC gateway configuration (v1) to the new v2 API.
In the legacy setup, DHCP and DHCP reservations are managed with dedicated resources and referenced in the gateway network.
In v2, the public gateway is migrated to use IPAM (IP Address Management) mode.
In 2023, DHCP functionality was moved from Public Gateways to Private Networks, DHCP resources are now no longer needed.

Note:
During migration, you need to trigger the migration call by setting the `migrate_to_v2` flag on your public gateway resource.
You can do this via the Terraform configuration or by using the Scaleway CLI/Console.

## Prerequisites

### Ensure the Latest Provider Version

Ensure your Scaleway provider is updated to at least version `2.52.0`.

```hcl
terraform {
  required_providers {
    scaleway = {
      source  = "scaleway/scaleway"
      version = "~> v2.52.0"
    }
  }
}
```

### Legacy Configuration

A typical legacy configuration might look like this:

```hcl
resource "scaleway_vpc" "main" {
   name = "foo"
}

resource "scaleway_vpc_private_network" "main" {
   name   = "bar"
   vpc_id = scaleway_vpc.main.id
}

resource "scaleway_vpc_public_gateway_ip" "main" {
}

resource "scaleway_vpc_public_gateway" "main" {
   name  = "foobar"
   type  = "VPC-GW-S"
   ip_id = scaleway_vpc_public_gateway_ip.main.id
}

resource "scaleway_vpc_public_gateway_dhcp" "main" {
   subnet = "192.168.1.0/24"
}

resource "scaleway_instance_server" "main" {
   image = "ubuntu_focal"
   type  = "DEV1-S"
}

resource "scaleway_instance_private_nic" "main" {
   server_id          = scaleway_instance_server.main.id
   private_network_id = scaleway_vpc_private_network.main.id
}

resource "scaleway_vpc_gateway_network" "main" {
   gateway_id         = scaleway_vpc_public_gateway.main.id
   private_network_id = scaleway_vpc_private_network.main.id
   dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
   cleanup_dhcp       = true
   enable_masquerade  = true
}

resource "scaleway_vpc_public_gateway_dhcp_reservation" "main" {
   gateway_network_id = scaleway_vpc_gateway_network.main.id
   mac_address        = scaleway_instance_private_nic.main.mac_address
   ip_address         = "192.168.1.1"
}
```

### Triggering Migration

Before updating your configuration, you must trigger the migration on the public gateway resource. For example, add the `migrate_to_v2` flag:

```hcl
resource "scaleway_vpc_public_gateway" "main" {
  name          = "foobar"
  type          = "VPC-GW-S"
  ip_id         = scaleway_vpc_public_gateway_ip.main.id
  migrate_to_v2 = true
}
```

This call updates the gateway on the API side from v1 to v2. The DHCP configuration and reservations remain intact, but the underlying resource is now managed using v2.

### Updated Configuration

After triggering the migration, update your Terraform configuration as follows:

1. **Remove the DHCP and DHCP Reservation Resources**

    Since DHCP functionality is built directly into Private Networks, you no longer need the DHCP configuration resources. Delete the following from your config:

    `scaleway_vpc_public_gateway_dhcp`
    `scaleway_vpc_public_gateway_dhcp_reservation`

2. **Update the Gateway Network**

    Replace the DHCP related attributes with an `ipam_config` block. For example

    ```hcl
    resource "scaleway_vpc_gateway_network" "main" {
      gateway_id         = scaleway_vpc_public_gateway.main.id
      private_network_id = scaleway_vpc_private_network.main.id
      ipam_config {
        push_default_route = false
      }
      enable_masquerade = true
    }
    ```

### Using the IPAM Datasource and Resource for Reservations

After migrating your public gateway to v2, you no longer manage DHCP reservations with dedicated resources.
Instead, you remove the legacy DHCP reservation resource and switch to using IPAM to manage your IPs.

1. **Retrieve an Existing IP with the IPAM Datasource**  
   If you have already reserved an IP (for example, via your legacy configuration), even after deleting the DHCP reservation resource the IP is still available. You can reference it using the `scaleway_ipam_ip` datasource. For instance:

   ```hcl
   data "scaleway_ipam_ip" "existing" {
     mac_address = scaleway_instance_private_nic.main.mac_address
     type        = "ipv4"
   }
   ```

   You can now use data.scaleway_ipam_ip.existing.id in your configuration to reference the reserved IP.

2. **Book New IPs Using the IPAM IP Resource**
   If you need to reserve new IPs, use the scaleway_ipam_ip resource. This resource allows you to explicitly book an IP from your private network. For example:

   ```hcl
   resource "scaleway_ipam_ip" "new_ip" {
     address = "172.16.64.7"
     source {
       private_network_id = scaleway_vpc_private_network.pn01.id
     }
   }
   ```

3. **Attach the Reserved IP to Your Resources**

   Once you have your IP—whether retrieved via the datasource or booked as a new resource—you can attach it to your server’s private NIC:

   ```hcl
   resource "scaleway_instance_private_nic" "pnic01" {
     private_network_id = scaleway_vpc_private_network.main.id
     server_id          = scaleway_instance_server.main.id
     ipam_ip_ids        = [scaleway_ipam_ip.new_ip.id]
   }
   ```
