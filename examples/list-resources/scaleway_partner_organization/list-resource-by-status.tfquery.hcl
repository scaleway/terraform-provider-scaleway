# List organizations managed by the partner filtered by status
list "scaleway_partner_organization" "by_status" {
  provider = scaleway

  config {
    status = "opened"
  }
}
