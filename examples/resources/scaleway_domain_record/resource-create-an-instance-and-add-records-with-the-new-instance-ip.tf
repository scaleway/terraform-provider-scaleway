### Create an Instance and add records with the new Instance IP

variable "project_id" {
  type        = string
  description = "Your project ID."
}

variable "dns_zone" {
  type        = string
  description = "The DNS Zone used for testing records."
}

resource "scaleway_instance_ip" "public_ip" {
  project_id = var.project_id
}

resource "scaleway_instance_server" "web" {
  project_id = var.project_id
  type       = "DEV1-S"
  image      = "ubuntu_jammy"
  tags       = ["front", "web"]
  ip_id      = scaleway_instance_ip.public_ip.id

  root_volume {
    size_in_gb = 20
  }
}

resource "scaleway_domain_record" "web_A" {
  dns_zone = var.dns_zone
  name     = "web"
  type     = "A"
  data     = scaleway_instance_server.web.public_ip
  ttl      = 3600
}

resource "scaleway_domain_record" "web_cname" {
  dns_zone = var.dns_zone
  name     = "www"
  type     = "CNAME"
  data     = "web.${var.dns_zone}."
  ttl      = 3600
}

resource "scaleway_domain_record" "web_alias" {
  dns_zone = var.dns_zone
  name     = ""
  type     = "ALIAS"
  data     = "web.${var.dns_zone}."
  ttl      = 3600
}
