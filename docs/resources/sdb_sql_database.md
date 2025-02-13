---
subcategory: "Databases"
page_title: "Scaleway: scaleway_sdb_sql_database"
---

# Resource: scaleway_sdb_sql_database

The `scaleway_sdb_sql_database` resource allows you to create and manage databases for Scaleway Serverless SQL Databases.

Refer to the Serverless SQL Databases [documentation](https://www.scaleway.com/en/docs/serverless-sql-databases/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-databases/) for more information.

## Example Usage

### Basic

```hcl
resource scaleway_sdb_sql_database "database" {
  name = "my-database"
  min_cpu = 0
  max_cpu = 8
}
```

### With IAM Application

This example creates an [IAM application](https://www.scaleway.com/en/docs/iam/concepts/#application) and an [API secret key](https://www.scaleway.com/en/docs/iam/how-to/create-api-keys/) used to connect to the database.

-> **Note:** For more information, see [How to connect to a Serverless SQL Database](https://www.scaleway.com/en/docs/serverless-sql-databases/how-to/connect-to-a-database/)

```hcl
data scaleway_account_project "default" {
  name = "default"
}

resource scaleway_iam_application "app" {
  name = "my app"
}

resource scaleway_iam_policy "db_access" {
  name = "my policy"
  description = "gives app access to serverless database in project"
  application_id = scaleway_iam_application.app.id
  rule {
    project_ids = [data.scaleway_account_project.default.id]
    permission_set_names = ["ServerlessSQLDatabaseReadWrite"]
  }
}

resource scaleway_iam_api_key "api_key" {
  application_id = scaleway_iam_application.app.id
}

resource scaleway_sdb_sql_database "database" {
  name = "my-database"
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
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the database (e.g. `my-new-database`).

    ~> **Important:** Updates to the `name` argument will recreate the database.

- `min_cpu` - (Optional) The minimum number of CPU units for your database. Defaults to 0.
- `max_cpu` - (Optional) The maximum number of CPU units for your database. Defaults to 15.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the resource exists.

## Attributes Reference

The `scaleway_sdb_sql_database` resource exports certain attributes once the Serverless SQL Database is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all the arguments above, the following attributes are exported:

- `id` - The unique identifier of the database, which is of the form `{region}/{id}` e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `endpoint` - The endpoint of the database.

## Import

Serverless SQL Databases can be imported using the `{region}/{id}`, as shown below:

```bash
terraform import scaleway_sdb_sql_database.database fr-par/11111111-1111-1111-1111-111111111111
```
