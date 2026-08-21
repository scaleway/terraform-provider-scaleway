# List file systems filtered by organization ID
list "scaleway_file_filesystem" "by_organization" {
  provider = scaleway

  config {
    regions         = ["*"]
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
