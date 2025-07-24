---
page_title: "Using Backend Guide"
---

# Terraform Backend

This page describes how to configure a backend by adding the backend block to your configuration with the Terraform Scaleway Provider.

Terraform provides the option to set up a [“backend”](https://developer.hashicorp.com/terraform/language/backend) of the `state` data files.

This option allows you to handle the state and the way certain operations are executed.

Backends can store the state remotely and protect it with locks to prevent corruption;
it makes it possible for a team to work with ease, or, for instance, to run Terraform within a pipeline.

## Create your database

You can create your database resource using terraform itself .

If you have already one database running you can step over to [Configuring your Connection string](#configuring-the-postgresql-connection-string)

```hcl
terraform {
  required_providers {
    scaleway = {
      source  = "scaleway/scaleway"
      version = "~> 2.2.8"
    }
  }
}

provider "scaleway" {
  # ...
}

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# CREATE AN DATABASE INSTANCE TO USE IT AS A TERRAFORM BACKEND
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

resource "scaleway_rdb_database" "database" {
  name        = "your-database"
  instance_id = scaleway_rdb_instance.main.id
}

resource scaleway_rdb_instance main {
  name           = "your-backend-db"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-11"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  tags           = ["terraform-backend1"]
}
```

and deploy it:

```shell
terraform plan -out "planfile" ; terraform apply -input=false -auto-approve "planfile"
```

## Configuring the PostgreSQL Connection String

We choose to set our environment variable for the connection string for this guide. Please check the [secret section](#secrets) for more details.

```shell
export PG_CONN_STR=postgres://<user>:<pass>@localhost:<port>/terraform_backend?sslmode=disable
```

## Secrets

Hashicorp offers several methods to keep your secrets. Please check the Terraform [partial configuration](https://developer.hashicorp.com/terraform/language/backend#partial-configuration) for this topic.

## Create your infrastructure with the Scaleway provider

```hcl
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# CREATE AN BACKEND TYPE PG
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
terraform {
  backend "pg" {
    # Please use a better approach with the flag -backend-config=PATH or a Vault configuration
    conn_str = "postgres://user:pass@db.example.com/terraform_backend"
  }
}

# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# CREATE YOUR INFRASTRUCTURE
# ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
resource "scaleway_instance_server" "main" {
  name        = "my-instance"
  type        = "DEV1-S"
  image       = "debian_bullseye"
  enable_ipv6 = false
}

# the rest of your configuration and resources to deploy
```

Check your database `schema`. e.g:

```sql
rdb=> SELECT * FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';
       schemaname       | tablename | tableowner | tablespace | hasindexes | hasrules | hastriggers | rowsecurity
------------------------+-----------+------------+------------+------------+----------+-------------+-------------
 terraform_remote_state | states    | my_initial_user |            | t          | f        | f           | f
```

After running terraform `apply`, to terraform.tfstate on the database will look something like this:

```text
rdb=> SELECT * FROM information_schema.columns
WHERE table_schema = 'terraform_remote_state'
AND TABLE_NAME = 'states';
 id |  name   | data
----+---------+----------------------------------------------------------
  1 | default | {                                                       +
    |         |   "version": 4,                                         +
    |         |   "serial": 0,                                          +
    |         |   "lineage": "07a1e05b-3cba-438a-0c70-3ec5e73d4baf",    +
    |         |   "outputs": {},                                        +
    |         |   "resources": [                                        +
    |         |     {
    
    ....
```

## Multiple Workplaces

You can configure several `states` on your database using a different `schema_name`.

Then workspaces are appended to that key to generate a separate state for each workspace.
Since tracking of the workspaces is in the table inside PostgreSQL, we need to separate the different states we want to track.

We can do that in one of two ways: separate databases or separate schemas.

```hcl
terraform {
  # Omitted

  backend "pg" {
    schema_name = "other_state"
  }
}
```

## Migrating the state

Considering you have already running infrastructure you want to use the `backend` option.

All we need to do is initialize Terraform passing the backend configuration.

Terraform should ask if you want to migrate from local to the new remote backend.

Answer the prompt `yes`, and your state will migrate.

```shell
$ terraform init -backend-config="conn_str=${PG_CONN_STR}" -migrate-state
```

## What about locking?

Most of the remote [backends](https://developer.hashicorp.com/terraform/language/backend#backend-types) natively support locking. To run terraform apply, Terraform will automatically acquire a lock;
if someone else is already running apply, they will already have the lock, and you will have to wait.
You can run apply with the `-lock-timeout=<TIME>` parameter to tell Terraform to wait up to TIME for a lock to be released (e.g., `-lock-timeout=10m` will wait for 10 minutes).

The Lock method prevents opening the state file while already in use.

## Share configuration

You can also share the configuration using the different [data sources](https://www.terraform.io/language/state/remote-state-data).
This is useful when working on the same infrastructure or the same team.

```hcl
data "scaleway_rdb_instance" "mybackend" {
  name = "your-database-name"
}
```

## Using s3 compatible backend

You can use scaleway object-storage bucket as an S3 compatible backend to store your terraform tfstate file just like you would do with AWS s3

### by using hardcoded credetials (NOT RECOMENDED)
```
terraform {
  backend "s3" {
    bucket                      = "scw-bucket"
    key                         = "terraform.tfstate"
    region                      = "fr-par"
    endpoint                    = "https://s3.fr-par.scw.cloud"
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_s3_checksum            = true
    access_key                  = XXXXXXXXXXX
    secert_key                  = YYYYYYYYYYY
  }
}
```

### by using credetials environment variables
```
$ export SCW_ACCESS_KEY="XXXXXXXXXXX"
$ export SCW_SECRET_KEY="YYYYYYYYYYY"
```
and this simple backend code
```
terraform {
  backend "s3" {
    bucket                      = "scw-bucket"
    key                         = "terraform.tfstate"
    region                      = "fr-par"
    endpoint                    = "https://s3.fr-par.scw.cloud"
  }
}
```

### and with the shared congfiguration file ? 
scaleway scw cli privide you with a credential file 
>~/$HOME/.config/scw/config.yaml

to generate your credential fil you can run scw init at first run or login for every new key pairs generated
```
scw login
```

it wil generat scw shared configuration fil folowinng this format
```
profiles:
  myProfile1:
    access_key: XXXXXXXXXXX
    secret_key: YYYYYYYYYYY
    default_organization_id: example-org-id-zzzzzzzzzz
    default_project_id: example-org-id-zzzzzzzzzz
    default_zone: fr-par-1
    default_region: fr-par
    api_url: https://api.scaleway.com
    insecure: false
```

actualy terraform backend "s3" is not aware of any other kind of s3 compatible bucket and is by default assuming you ar using aws's S3 service

so in order to read scw ccredentials, do not try to use `profile = myProfile1` it will not work, unless you copy scw credentials into aws shared configuration file

>~/$HOME/.aws/credentials

```
[scaleway_profile]
aws_access_key_id = XXXXXXXXXXX
aws_secret_access_key = YYYYYYYYYYY
```
then in your tf backend bloc use this profile as if it waf a aws backend `profile = "scaleway_rofile"`

```
terraform {
  backend "s3" {
    bucket                      = "scw-bucket"
    key                         = "terraform.tfstate"
    region                      = "fr-par"
    endpoint                    = "https://s3.fr-par.scw.cloud"
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_s3_checksum            = true
    profile                     = "scaleway_rofile"
  }
}
```
now run terraform init and backend s3 should be able to use scaleway object storage instead of aws s3