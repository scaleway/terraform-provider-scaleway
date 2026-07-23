### Basic

# Find VPCs that share the same tags
data "scaleway_vpcs" "my_key" {
  tags = ["tag"]
}

# Find VPCs by name and region
data "scaleway_vpcs" "my_key" {
  name   = "tf-vpc-datasource"
  region = "nl-ams"
}
