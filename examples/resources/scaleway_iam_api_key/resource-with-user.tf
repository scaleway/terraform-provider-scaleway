### With user

resource "scaleway_iam_user" "main" {
  email = "test@test.com"
}

resource "scaleway_iam_api_key" "main" {
  user_id     = scaleway_iam_user.main.id
  description = "a description"
}
