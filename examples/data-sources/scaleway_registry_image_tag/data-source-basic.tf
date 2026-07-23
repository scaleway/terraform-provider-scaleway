### Example Usage

# Get info by tag ID
data "scaleway_registry_image_tag" "my_image_tag" {
  tag_id = "11111111-1111-1111-1111-111111111111"
}

# Get info by name and image_id
data "scaleway_registry_image_tag" "my_image_tag" {
  name     = "my-tag-name"
  image_id = "22222222-2222-2222-2222-222222222222"
}
