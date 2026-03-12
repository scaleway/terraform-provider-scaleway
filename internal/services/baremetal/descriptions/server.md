Creates and manages Scaleway Compute Baremetal servers. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/elastic-metal/).

-> **Security Best Practice:**
For enhanced security, we recommend using the [`password_wo` and `service_password_wo` write-only arguments](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) instead of the regular `password` and `service_password` arguments. This ensures your sensitive data is never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in Terraform 1.11.0 and later.
