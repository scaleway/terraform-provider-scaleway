# When using mutable images (e.g., `latest` tag), you can use the `scaleway_registry_image_tag` data source along 
# with the `registry_sha256` argument to trigger container redeployments when the image is updated.

# Ideally, you would create the namespace separately.
# For demonstration purposes, this example assumes the "nginx:latest" image is already available
# in the referenced namespace.
resource "scaleway_registry_namespace" "main" {
  name = "some-unique-name"
}

data "scaleway_registry_image" "nginx" {
  namespace_id = scaleway_registry_namespace.main.id
  name         = "nginx"
}

data "scaleway_registry_image_tag" "nginx_latest" {
  image_id = data.scaleway_registry_image.nginx.id
  name     = "latest"
}

resource "scaleway_container_namespace" "main" {
  name = "my-container-namespace"
}

resource "scaleway_container" "main" {
  name            = "nginx-latest"
  namespace_id    = scaleway_container_namespace.main.id
  registry_image  = "${scaleway_registry_namespace.main.endpoint}/nginx:latest"
  registry_sha256 = data.scaleway_registry_image_tag.nginx_latest.digest
  port            = 80
  deploy          = true
}

# Using this configuration, whenever the `latest` tag of the `nginx` image is updated, the `registry_sha256` will change, triggering a redeployment of the container with the new image.
