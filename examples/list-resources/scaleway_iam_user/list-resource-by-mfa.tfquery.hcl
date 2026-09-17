// List users filtered by MFA status
list "scaleway_iam_user" "by_mfa" {
  provider = scaleway

  config {
    mfa = true
  }
}
