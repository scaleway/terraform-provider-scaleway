The scaleway_inference_model resource allows you to upload and manage inference models in the Scaleway Inference ecosystem. Once registered, a model can be used in any scaleway_inference_deployment resource.

-> **Security Best Practice:**
For enhanced security, we recommend using the [`secret_wo` write-only argument](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) instead of the regular `secret` argument. This ensures your sensitive credentials are never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later.
