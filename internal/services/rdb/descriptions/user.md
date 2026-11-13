The [`scaleway_rdb_user`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/rdb_user) resource creates and manages database users.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/).

-> **Security Best Practice:**
For enhanced security, we recommend using the [`password_wo` write-only argument](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) instead of the regular `password` argument. This ensures your sensitive credentials are never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in Terraform 1.11.0 and later.
