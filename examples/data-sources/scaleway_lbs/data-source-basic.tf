### Basic

# Find LBs by name
data "scaleway_lbs" "my_key" {
  name = "foobar"
}

# Find LBs by name and zone
data "scaleway_lbs" "my_key" {
  name = "foobar"
  zone = "fr-par-2"
}

# Find LBs that share the same tags
data "scaleway_lbs" "lbs_by_tags" {
  tags = ["a tag"]
}
