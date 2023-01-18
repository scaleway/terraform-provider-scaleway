---
page_title: "Using Scaleway Messaging and Queuing service with NATS Terraform provider"
description: |-
Using Scaleway Messaging and Queuing service with NATS Terraform provider
---

# How to use Scaleway Messaging and Queuing config

In this guide you'll learn how to deploy Scaleway Messaging and Queuing config and use the NATS Jetstream provider with,
which is a plugin for Terraform that allows you to provision and manage NATS Jetstream resources.

## Prerequisites

* First, you will need to set up a new Terraform configuration file
  with [Scaleway](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/mnq_namespace)
  and [Jetstream provider](https://registry.terraform.io/providers/nats-io/jetstream/latest/docs/guides/setup).

```hcl
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
    jetstream = {
      source  = "nats-io/jetstream"
      version = "~> 0.0.34"
    }
  }
}

provider "scaleway" {
  region     = "fr-par"
  access_key = "<SCW_ACCESS_KEY>"
  secret_key = "<SCW_SECRET_KEY>"
  project_id = "<SCW_DEFAULT_PROJECT_ID>"
}
```

* Next, you should create a Messaging and Queuing namespace. Check our example below.

```hcl
resource "scaleway_mnq_namespace" "main" {
  name     = "mnq-ns"
  protocol = "nats"
}
```

* At this point you can configure the [NATS CLI](https://docs.nats.io/using-nats/nats-tools/nats_cli) with your
  endpoint.

```shell
nats context save example --server nats://nats.mnq.fr-par.scw.cloud:4222 --description 'Prod.Net Server'
```

* NATS let you use `contexts` that you can store and easily select. Check more
  details [here](https://docs.nats.io/using-nats/nats-tools/nats_cli#configuration-contexts).

```shell
NATS Configuration Context "example"

      Description: Prod.Net Server
      Server URLs: nats://nats.mnq.fr-par.scw.cloud:4222
             Path: /Your/path/context/example.json
```

* Before to create your credentials, be aware that despite being an
  ongoing [issue](https://github.com/hashicorp/terraform/issues/516) since 2014, secrets stored in `terraform.tfstate`
  remain in plain text. While there are methods to remove secrets from state files, they are unreliable and may not
  function properly with updates to Terraform. It is not recommended to use these workarounds.

---
  Otherwise, the official update on December3, 2020:
  Terraform 0.14 has added the ability to mark variables as sensitive, which helps keep them out of your logs, so you should add `sensitive = true` to variables!
---

* You can create a Credential easily using the Scaleway provider.
  Check our example below and more about authenticating with a Credentials
  File [here](https://docs.nats.io/using-nats/developer/connecting/creds)

```hcl
resource "scaleway_mnq_credential" "main" {
  name         = "creed-ns"
  namespace_id = scaleway_mnq_namespace.main.id
}
```

* Try to select your configuration using the NATS CLI:
```shell
nats context select
? Select a Context  [Use arrows to move, type to filter]
> example
```

```shell
NATS Configuration Context "example"

      Description: Prod.Net Server
      Server URLs: nats://nats.mnq.fr-par.scw.cloud:4222
      Credentials: /Your/path/secret/admin.creds (OK)
             Path: /Your/path/context/example.json
```

* Finally configuring the provider with the server and credentials for the NATS Jetstream service.

```hcl
provider "jetstream" {
  servers     = scaleway_mnq_namespace.manin.endpoint
  credentials = "path/ngs_stream_admin.creds"
}
```
