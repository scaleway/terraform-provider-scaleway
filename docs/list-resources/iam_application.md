---
page_title: "Scaleway: scaleway_iam_application"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM Applications.
---

# Resource: scaleway_iam_application

Lists Scaleway IAM Applications.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
# List all applications in an organization
list "scaleway_iam_application" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
# List applications filtered by application IDs
list "scaleway_iam_application" "by_application_ids" {
  provider = scaleway

  config {
    application_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
```

```terraform
# List applications filtered by editable status
list "scaleway_iam_application" "by_editable" {
  provider = scaleway

  config {
    editable = true
  }
}
```

```terraform
# List applications filtered by name
list "scaleway_iam_application" "by_name" {
  provider = scaleway

  config {
    name = "my-application"
  }
}
```

```terraform
# List applications filtered by tag
list "scaleway_iam_application" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `organization_id` - (Optional) Organization ID to filter for. If not specified, the provider default organization is used.
- `name` - (Optional) Name of the application to filter for.
- `tag` - (Optional) Tag to filter for.
- `editable` - (Optional) Filter by editable status.
- `application_ids` - (Optional) Filter applications by application IDs.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Application:

- `id` - The ID of the application.
- `name` - The name of the application.
- `description` - The description of the application.
- `created_at` - The date and time of the creation of the application.
- `updated_at` - The date and time of the last update of the application.
- `organization_id` - The organization ID the application belongs to.
- `tags` - The tags associated with the application.
- `editable` - Defines whether or not the application is editable.
- `deletable` - Defines whether or not the application is deletable.
- `managed` - Defines whether or not the application is managed.
- `nb_api_keys` - Number of API keys attributed to the application.
