---
page_title: "Custom Service Endpoints Guide"
---

# Custom Service Endpoints Configuration

The Terraform Scaleway Provider configuration can be customized to connect to
non-default service endpoints and compatible solutions. This may be useful for
environments with specific compliance requirements.

~> **Note:** Support for connecting the Terraform Scaleway Provider with custom
endpoints and compatible solutions is offered as best effort. Individual
Terraform resources may require compatibility updates to work in certain
environments.

## Getting started with custom endpoints

To configure endpoints on the provider, set the values in the `endpoints` block in
the `provider` declarations, e.g.,

```hcl
provider "scaleway" {
  # ... potentially other provider configuration ...
  
  endpoints {
    s3 = "http://localhost:4572"
  }
}
```

~> **Important**: When using `localhost` as an S3 endpoint, make sure to enable
[`s3_use_path_style`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs#s3_use_path_style-1),
so that your buckets are accessed without DNS errors.

## Available Endpoint Customizations

| Service        | Provider Parameter |
|----------------|--------------------|
| Object Storage | `s3`               |
