### With IAM Application

data "scaleway_account_project" "default" {
  name = "default"
}

resource "scaleway_iam_application" "app" {
  name = "my app"
}

resource "scaleway_iam_policy" "db_access" {
  name           = "my policy"
  description    = "gives app access to serverless database in project"
  application_id = scaleway_iam_application.app.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["ServerlessSQLDatabaseReadWrite"]
  }
}

resource "scaleway_iam_api_key" "api_key" {
  application_id = scaleway_iam_application.app.id
}

resource "scaleway_sdb_sql_database" "database" {
  name    = "my-database"
  min_cpu = 0
  max_cpu = 8
}

output "database_connection_string" {
  // Output as an example, you can give this string to your application
  value = format("postgres://%s:%s@%s",
    scaleway_iam_application.app.id,
    scaleway_iam_api_key.api_key.secret_key,
    trimprefix(scaleway_sdb_sql_database.database.endpoint, "postgres://"),
  )
  sensitive = true
}
