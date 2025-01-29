---
subcategory: "Load Balancers"
page_title: "Scaleway: scaleway_lb_certificate"
---

# scaleway_lb_certificate

Get information about Scaleway Load Balancer certificates.

This data source can prove useful when a module accepts a Load Balancer certificate as an input variable and needs to, for example, determine the security of a certificate for the frontend associated with your domain.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/load-balancer/how-to/add-certificate/) or [API documentation](https://www.scaleway.com/en/developers/api/load-balancer/zoned-api/#path-certificate).

## Examples

### Let's Encrypt

```hcl
resource scaleway_lb_ip main {
}

resource scaleway_lb main {
    ip_id = scaleway_lb_ip.main.id
    name = "data-test-lb-cert"
    type = "LB-S"
}

resource scaleway_lb_certificate main {
    lb_id = scaleway_lb.main.id
    name = "data-test-lb-cert"
    letsencrypt {
        common_name = "${replace(scaleway_lb.main.ip_address, ".", "-")}.lb.${scaleway_lb.main.region}.scw.cloud"
    }
}

data "scaleway_lb_certificate" "byID" {
    certificate_id = "${scaleway_lb_certificate.main.id}"
}

data "scaleway_lb_certificate" "byName" {
    name = "${scaleway_lb_certificate.main.name}"
    lb_id = "${scaleway_lb.main.id}"
}
```

## Arguments Reference

The following arguments are supported:

- `certificate_id` - (Optional) The certificate ID.
    - Only one of `name` and `certificate_id` should be specified.

- `name` - (Optional) The name of the Load Balancer certificate.
    - When using a certificate `name` you should specify the `lb-id`

- `lb_id` - (Required) The Load Balancer ID this certificate is attached to.

## Attributes Reference

See the [Load Balancer certificate resource](../resources/lb_certificate.md) for details on the returned attributes - they are identical.