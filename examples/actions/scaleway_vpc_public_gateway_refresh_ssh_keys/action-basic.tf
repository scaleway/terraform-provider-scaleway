resource "scaleway_vpc_public_gateway" "main" {
  name            = "tf-test-vpcgw-action-refresh-ssh-keys"
  type            = "VPC-GW-S"
  bastion_enabled = true

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_vpc_public_gateway_refresh_ssh_keys.main]
    }
  }
}

action "scaleway_vpc_public_gateway_refresh_ssh_keys" "main" {
  config {
    gateway_id = scaleway_vpc_public_gateway.main.id
    wait       = true
  }
}
