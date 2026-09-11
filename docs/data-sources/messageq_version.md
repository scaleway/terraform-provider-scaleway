---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_version"
---

# Data Source: scaleway_messageq_version

Gets information about an available MessageQ version.

## Example Usage

```terraform
data "scaleway_messageq_version" "latest" {
  name = "latest"
}

data "scaleway_messageq_version" "specific" {
  name = "4.0"
}
```

## Argument Reference

- `name` - (Required) The MessageQ version name. Use `latest` to retrieve the most recent available non-disabled version.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The region in which the version is available.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the version in the `{region}/{version}` format.
- `version` - The MessageQ version string.
- `end_of_life` - The end-of-life date of the version (RFC 3339 format).
- `disabled` - Whether the version is disabled.
- `beta` - Whether the version is in beta.
