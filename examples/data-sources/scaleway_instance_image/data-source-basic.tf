# Get info by image name
data "scaleway_instance_image" "my_image" {
  name = "my-image-name"
}

# Get info by image id
data "scaleway_instance_image" "my_image" {
  image_id = "11111111-1111-1111-1111-111111111111"
}
