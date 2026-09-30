### Create a policy for an organization's project

provider "scaleway" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

data "scaleway_account_project" "default" {
  name = "default"
}

resource "scaleway_iam_application" "app" {
  name = "my app"
}

resource "scaleway_iam_policy" "object_read_only" {
  name           = "my policy"
  description    = "gives app readonly access to object storage in project"
  application_id = scaleway_iam_application.app.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["ObjectStorageReadOnly"]
  }
}
