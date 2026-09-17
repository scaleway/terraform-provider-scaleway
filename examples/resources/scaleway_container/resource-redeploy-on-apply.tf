data "scaleway_registry_namespace" "main" {
  name = "my-registry"
}

data "scaleway_registry_image" "main" {
  namespace_id = data.scaleway_registry_namespace.main.id
  name         = "nginx-1-29-2-alpine"
}

resource "scaleway_container_namespace" "main" {}

resource "scaleway_container" "main" {
  name         = "my-container"
  namespace_id = scaleway_container_namespace.main.id
  image        = "${data.scaleway_registry_namespace.main.endpoint}/${data.scaleway_registry_image.main.name}:${data.scaleway_registry_image.main.tags[0]}"
  port         = 80

  # At every update, timestamp() will trigger a change and redeploy the container, even though nothing else has changed.
  registry_sha256 = timestamp()
}
