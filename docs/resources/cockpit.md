---
page_title: "Scaleway: scaleway_cockpit"
description: |-
    Manages Scaleway Cockpits.
---

# scaleway_cockpit

Creates and manages Scaleway Cockpit.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/).

## Example Usage

```hcl
// Create the cockpit in the default project
resource "scaleway_cockpit" "main" {}
```

```hcl
// Create the cockpit in a specific project
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
}
```

## Arguments Reference

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the domain is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `endpoints` - Endpoints
    - `metrics_url` - The metrics URL
    - `logs_url` - The logs URL
    - `alertmanager_url` - The alertmanager URL
    - `grafana_url` - The grafana URL
