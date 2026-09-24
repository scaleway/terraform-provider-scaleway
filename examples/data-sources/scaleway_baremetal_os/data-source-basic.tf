## Basic

# Get info by os name and version
data "scaleway_baremetal_os" "by_name" {
  name    = "Ubuntu"
  version = "20.04 LTS (Focal Fossa)"
}

# Get info by os id
data "scaleway_baremetal_os" "by_id" {
  os_id = "03b7f4ba-a6a1-4305-984e-b54fafbf1681"
}
