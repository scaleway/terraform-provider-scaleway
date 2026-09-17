resource "scaleway_container_namespace" "main" {}

resource "scaleway_container" "app" {
  name         = "app"
  namespace_id = scaleway_container_namespace.main.id
  image        = "nginx:latest"
  port         = 80
  privacy      = "public"
  protocol     = "http1"
}

resource "scaleway_domain_record" "app" {
  dns_zone = "scaleway-terraform.com"
  name     = "subdomain"
  type     = "CNAME"
  data     = format("%s.", trimprefix("${scaleway_container.app.public_endpoint}", "https://")) // Trailing dot is important in CNAME
  ttl      = 3600
}

resource "scaleway_container_domain" "app" {
  container_id = scaleway_container.app.id
  hostname     = "${scaleway_domain_record.app.name}.${scaleway_domain_record.app.dns_zone}"
}
