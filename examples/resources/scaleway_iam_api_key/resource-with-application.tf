### With application

resource "scaleway_iam_application" "ci_cd" {
  name = "My application"
}

resource "scaleway_iam_api_key" "main" {
  application_id = scaleway_iam_application.main.id
  description    = "a description"
}
