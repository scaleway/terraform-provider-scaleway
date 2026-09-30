### Deploy an SSH key in all your organization's projects

data "scaleway_account_projects" "all" {}

resource "scaleway_account_ssh_key" "main" {
  name       = "main"
  public_key = local.public_key
  count      = length(data.scaleway_account_projects.all.projects)
  project_id = data.scaleway_account_projects.all.projects[count.index].id
}
