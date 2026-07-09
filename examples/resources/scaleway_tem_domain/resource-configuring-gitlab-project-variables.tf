### Configuring GitLab Project Variables


variable "domain_name" {
  type = string
}

data "scaleway_tem_domain" "my_domain" {
  name = var.domain_name
}

resource "gitlab_project_variable" "smtp_auth_user" {
  key   = "SMTP_AUTH_USER"
  value = data.scaleway_tem_domain.my_domain.smtps_auth_user
}

resource "gitlab_project_variable" "smtp_port" {
  key   = "SMTP_PORT"
  value = data.scaleway_tem_domain.my_domain.smtps_port
}

resource "gitlab_project_variable" "smtp_host" {
  key   = "SMTP_HOST"
  value = data.scaleway_tem_domain.my_domain.smtps_host
}
