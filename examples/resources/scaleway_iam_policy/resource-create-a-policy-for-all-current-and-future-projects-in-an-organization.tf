### Create a policy for all current and future projects in an organization

resource "scaleway_iam_application" "app" {
  name = "my app"
}

resource "scaleway_iam_policy" "object_read_only" {
  name           = "my policy"
  description    = "gives app readonly access to object storage in project"
  application_id = scaleway_iam_application.app.id
  rule {
    organization_id      = scaleway_iam_application.app.organization_id
    permission_set_names = ["ObjectStorageReadOnly"]
  }
}
