## Basic

# Get info by os name and version
data "scaleway_apple_silicon_os" "by_name" {
  name = "devos-sequoia-15.6"
}

# Get info by os id
data "scaleway_apple_silicon_os" "by_id" {
  os_id = "cafecafe-5018-4dcd-bd08-35f031b0ac3e"
}
