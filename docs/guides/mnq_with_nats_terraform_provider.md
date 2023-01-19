---
page_title: "Using Scaleway Messaging and Queuing service with NATS Terraform provider"
description: |-
Using Scaleway Messaging and Queuing service with NATS Terraform provider
---

# How to use Scaleway Messaging and Queuing config

In this guide you'll learn how to deploy Scaleway Messaging and Queuing config and use the NATS Jetstream provider with,
which is a plugin for Terraform that allows you to provision and manage NATS Jetstream resources.

At the end of this guide, you will have a running NATS server and a configured NATS CLI to interact with it.

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

* Before creating your credentials, be aware that due to an
  ongoing [issue](https://github.com/hashicorp/terraform/issues/516) since 2014, secrets stored in `terraform.tfstate`
  remain in plain text. While there are methods to remove secrets from state files, they are unreliable and may not
  function properly with updates to Terraform. It is not recommended to use these workarounds.

---

* Otherwise, the official update on December3, 2020:
  Terraform 0.14 has added the ability to mark variables as sensitive, which helps keep them out of your logs, so you
  should add `sensitive = true` to variables!

---

* You can create Credentials easily using the Scaleway provider.
  Check our example below and more about authenticating with a Credentials
  File [here](https://docs.nats.io/using-nats/developer/connecting/creds)

```hcl
resource "scaleway_mnq_credential" "main" {
  name         = "creds-ns"
  namespace_id = scaleway_mnq_namespace.main.id
}
```

* At this point you have your Namespace and your Credential that means you have a running NATS Server ready to be used.

* Grab a copy of the  [NATS CLI](https://github.com/nats-io/jetstream/releases) and configure it with your
  endpoint. To be practical let's use `contexts` that you can store and easily select the relevant context. Check more
  details [here](https://docs.nats.io/using-nats/nats-tools/nats_cli#configuration-contexts).

```shell
nats context save example --server nats://nats.mnq.fr-par.scw.cloud:4222 --description 'Prod.Net Server'
```

* The output should look like this:

```shell
NATS Configuration Context "example"

      Description: Prod.Net Server
      Server URLs: nats://nats.mnq.fr-par.scw.cloud:4222
             Path: /Your/path/context/example.json
```

* Try to select your configuration using the NATS CLI:

```shell
nats context select
? Select a Context  [Use arrows to move, type to filter]
> example
```

* Finally, configure the CLI with the namespace credentials.

```shell
NATS Configuration Context "example"

      Description: Prod.Net Server
      Server URLs: nats://nats.mnq.fr-par.scw.cloud:4222
      Credentials: /Your/path/secret/admin.creds (OK)
             Path: /Your/path/context/example.json
```

* You are ready to use the
  NATS [JetStream Provider](https://registry.terraform.io/providers/nats-io/jetstream/latest/docs):

```hcl
provider "jetstream" {
  servers     = scaleway_mnq_namespace.manin.endpoint
  credentials = "path/ngs_stream_admin.creds"
  # credential_data = "<SCW_CREDENTIAL_AS_STRING>"
}
```
