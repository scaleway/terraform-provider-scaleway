### Use a project in provider configuration

provider "scaleway" {
  alias = "tmp"
}

resource "scaleway_account_project" "project" {
  provider = scaleway.tmp
  name     = "my_project"
}

provider "scaleway" {
  project_id = scaleway_account_project.project.id
}

resource "scaleway_instance_server" "server" { // Will use scaleway_account_project.project
  image = "ubuntu_jammy"
  type  = "PRO2-XXS"
}
