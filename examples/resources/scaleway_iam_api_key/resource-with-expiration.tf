### With expiration

resource "time_rotating" "rotate_after_a_year" {
  rotation_years = 1
}

resource "scaleway_iam_api_key" "main" {
  application_id = scaleway_iam_application.main.id
  expires_at     = time_rotating.rotate_after_a_year.rotation_rfc3339
}
