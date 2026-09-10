---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_certificate_authority"
---

# Data Source: scaleway_messageq_certificate_authority

Downloads the certificate authority (CA) of a MessageQ deployment for TLS clients.

## Example Usage

```terraform
resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-SHARED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

data "scaleway_messageq_certificate_authority" "main" {
  deployment_id = scaleway_messageq_deployment.main.id
}
```

## Argument Reference

- `deployment_id` - (Required) The ID of the MessageQ deployment.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The region in which the deployment exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source in the `{region}/{deployment_id}` format.
- `pem` - (Sensitive) PEM-encoded certificate authority content.
