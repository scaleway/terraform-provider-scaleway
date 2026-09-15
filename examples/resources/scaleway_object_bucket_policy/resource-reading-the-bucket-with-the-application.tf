#### Reading the bucket with the application

data "scaleway_iam_application" "reading-app" {
  name = "reading-app"
}
resource "scaleway_iam_api_key" "reading-api-key" {
  application_id = data.scaleway_iam_application.reading-app.id
}

provider "scaleway" {
  access_key = scaleway_iam_api_key.reading-api-key.access_key
  secret_key = scaleway_iam_api_key.reading-api-key.secret_key
  alias      = "reading-profile"
}

data "scaleway_object_bucket" "bucket" {
  provider   = scaleway.reading-profile
  name       = "some-unique-name"
  depends_on = [scaleway_iam_api_key.reading-api-key]
}
