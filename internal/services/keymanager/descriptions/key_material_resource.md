Import externally generated key material into Key Manager to derive a new cryptographic key. The key's origin must be external.

-> **Security Best Practice:**
For enhanced security, we recommend using the [`key_material_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) and [`salt_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) write-only arguments instead of the regular `key_material` and `salt` arguments. This ensures your sensitive cryptographic material is never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in Terraform 1.11.0 and later.

-> **Note:** When using write-only arguments (`key_material_wo` and `salt_wo`), you must also provide the corresponding version fields (`key_material_wo_version` and `salt_wo_version`) to enable proper resource lifecycle management.
