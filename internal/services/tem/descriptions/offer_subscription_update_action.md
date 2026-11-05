The [`scaleway_tem_update_offer_subscription`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/actions/tem_update_offer_subscription) action subscribes a project to a Transactional Email offer, or updates an existing offer subscription (`essential` or `scale`).

There is no API to cancel an offer subscription. Use the [`scaleway_tem_offer_subscription`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/data-sources/tem_offer_subscription) data source to read the current subscription.

Refer to the Transactional Email [documentation](https://www.scaleway.com/en/docs/transactional-email/how-to/manage-tem-plans/) and [API documentation](https://www.scaleway.com/en/developers/api/transactional-email/offer-subscriptions) for more information.
