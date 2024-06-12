---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_plan"
---
# scaleway_cockpit_plan

The page provides documentation for `scaleway_cockpit_plan`, which is used to fetch details about a specific Scaleway Cockpit pricing plan by its name. This information can then be used to configure resources like `scaleway_cockpit`.

Find out more about [pricing plans](https://console.scaleway.com/cockpit/plans) in the Scaleway console.

## Fetch and associate a pricing plan to a Cockpit

The following command shows how to fetch information about the `premium` pricing plan and how to associate it with the Cockpit of your Scaleway default Project.

```hcl
data "scaleway_cockpit_plan" "premium" {
  name = "premium"
}

resource "scaleway_cockpit" "main" {
  plan = data.scaleway_cockpit_plan.premium.id
}
```

## Arguments reference

- `name` - This is a required argument that specifies the name of the pricing plan you want to retrieve information about.
