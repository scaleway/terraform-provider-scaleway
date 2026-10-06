# List block volumes filtered by organization ID
list "scaleway_block_volume" "by_organization" {
  provider = scaleway

  config {
    zones           = ["*"]
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
