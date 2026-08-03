### With applications

resource "scaleway_iam_application" "app" {}

resource "scaleway_iam_group" "with_app" {
  name = "iam_group_with_app"
  application_ids = [
    scaleway_iam_application.app.id,
  ]
  user_ids = []
}
