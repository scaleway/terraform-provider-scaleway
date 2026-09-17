resource "scaleway_account_project" "mail_project" {
  name = "mail-project"
}

resource "scaleway_mailbox_domain" "project_domain" {
  name       = "mail.example.com"
  project_id = scaleway_account_project.mail_project.id
}
