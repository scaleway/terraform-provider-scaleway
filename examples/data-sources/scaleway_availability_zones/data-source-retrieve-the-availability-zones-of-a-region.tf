## Retrieve the Availability Zones of a Region

# Get info by Region key
data "scaleway_availability_zones" "main" {
  region = "nl-ams"
}
