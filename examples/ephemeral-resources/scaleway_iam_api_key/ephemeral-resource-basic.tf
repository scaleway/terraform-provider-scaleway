ephemeral "scaleway_iam_api_key" "main" {
  application_id = scaleway_iam_application.main.id
  expires_at     = timeadd(timestamp(), "10h")
}
