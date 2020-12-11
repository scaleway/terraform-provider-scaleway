---
page_title: "Scaleway: scaleway_lb_certificate"
description: |-
  Manages Scaleway Load-Balancer Certificates.
---

# scaleway_lb_certificate

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancer Certificates. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples

### Let's Encrypt

```hcl
resource "scaleway_lb_certificate" "cert01" {
  lb_id = scaleway_lb.lb01.id
  name  = "cert1"

  letsencrypt {
    common_name = "example.org"
    subject_alternative_name = [
      "sub1.example.com",
      "sub2.example.com"
    ]
  }
}
```

### Custom Certificate

```hcl
resource "scaleway_lb_certificate" "cert01" {
  lb_id = scaleway_lb.lb01.id
  name  = "custom-cert"
  custom_certificate {
    certificate_chain = <<EOF
CERTIFICATE_CHAIN_CONTENTS
EOF
  }
}
```

## Arguments Reference

The following arguments are supported:

### Basic arguments

- `lb_id` - (Required) The load-balancer ID this certificate is attached to.

~> **Important:** Updates to `lb_id` will recreate the load-balancer certificate.

- `name` - (Optional) The name of the certificate backend.

- `letsencrypt` - (Optional) Configuration block for Let's Encrypt configuration. Only one of `letsencrypt` and `custom_certificate` should be specified.

    - `common_name` - (Required) Main domain of the certificate.

    - `subject_alternative_name` - (Optional) Array of alternative domain names.

~> **Important:** Updates to `letsencrypt` will recreate the load-balancer certificate.

- `custom_certificate` - (Optional) Configuration block for custom certificate chain. Only one of `letsencrypt` and `custom_certificate` should be specified.

    - `certificate_chain` - (Required) Full PEM-formatted certificate chain.

~> **Important:** Updates to `custom_certificate` will recreate the load-balancer certificate.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the loadbalancer certificate.
- `common_name` - Main domain of the certificate
- `subject_alternative_name` - The alternative domain names of the certificate
- `fingerprint` - The identifier (SHA-1) of the certificate
- `not_valid_before` - The not valid before validity bound timestamp
- `not_valid_after` - The not valid after validity bound timestamp
- `status` - Certificate status

## Additional notes

* Ensure that all domain names used in configuration are pointing to the load balancer IP.
  You can achieve this by creating a DNS record through terraform pointing to  `ip_address` property of `lb_beta` entity.
* In case there are any issues with the certificate, you will receive a `400` error from the `apply` operation.
  Use `export TF_LOG=DEBUG` to view exact problem returned by the api.
* Wildcards are not supported with Let's Encrypt yet.
