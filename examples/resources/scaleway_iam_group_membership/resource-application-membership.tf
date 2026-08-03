### Application Membership

resource "scaleway_iam_group" "group" {
  name                = "my_group"
  external_membership = true
}

resource "scaleway_iam_application" "app" {
  name = "my_app"
}

resource "scaleway_iam_group_membership" "member" {
  group_id       = scaleway_iam_group.group.id
  application_id = scaleway_iam_application.app.id
}
