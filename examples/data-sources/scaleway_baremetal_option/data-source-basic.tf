## Basic

# Get info by option name
data "scaleway_baremetal_option" "by_name" {
  name = "Remote Access"
}

# Get info by option id
data "scaleway_baremetal_option" "by_id" {
  option_id = "931df052-d713-4674-8b58-96a63244c8e2"
}
