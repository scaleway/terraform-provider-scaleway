package tem

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceOfferSubscription() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceOfferSubscriptionRead,
		SchemaFunc:  offerSubscriptionSchema,
	}
}

func offerSubscriptionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"organization_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "The organization_id the offer subscription is associated to",
			ValidateDiagFunc: verify.IsUUID(),
			ConflictsWith:    []string{"project_id"},
		},
		"offer_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the offer",
		},
		"subscribed_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of the subscription",
		},
		"cancellation_available_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Date and time of the end of the offer-subscription commitment",
		},
		"sla": {
			Type:        schema.TypeFloat,
			Computed:    true,
			Description: "Service Level Agreement percentage of the offer-subscription",
		},
		"max_domains": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Max number of domains that can be associated with the offer-subscription",
		},
		"max_dedicated_ips": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Max number of dedicated IPs that can be associated with the offer-subscription",
		},
		"max_webhooks_per_domain": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Max number of webhooks that can be associated with the offer-subscription",
		},
		"max_custom_blocklists_per_domain": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Max number of custom blocklists that can be associated with the offer-subscription",
		},
		"included_monthly_emails": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Number of emails included in the offer-subscription per month",
		},
	}
}

func DataSourceOfferSubscriptionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &tem.ListOfferSubscriptionsRequest{
		Region: region,
	}

	// ListOfferSubscriptionsRequest requires either an organization_id or a project_id.
	// When organization_id is set it takes precedence, otherwise fall back to the
	// project_id (from the configuration or the default project of the provider).
	if orgID, ok := d.GetOk("organization_id"); ok {
		req.OrganizationID = types.ExpandStringPtr(orgID)
	} else {
		projectID, _, err := meta.ExtractProjectID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		req.ProjectID = types.ExpandStringPtr(projectID)
	}

	offer, err := api.ListOfferSubscriptions(req, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if len(offer.OfferSubscriptions) == 0 {
		d.SetId("")

		return nil
	}

	offerSubscription := offer.OfferSubscriptions[0]
	d.SetId(regional.NewIDString(region, offerSubscription.ProjectID))
	_ = d.Set("project_id", offerSubscription.ProjectID)
	_ = d.Set("organization_id", offerSubscription.OrganizationID)
	_ = d.Set("region", region)
	_ = d.Set("offer_name", offerSubscription.OfferName)
	_ = d.Set("subscribed_at", types.FlattenTime(offerSubscription.SubscribedAt))
	_ = d.Set("cancellation_available_at", types.FlattenTime(offerSubscription.CancellationAvailableAt))
	_ = d.Set("sla", offerSubscription.SLA)
	_ = d.Set("max_domains", offerSubscription.MaxDomains)
	_ = d.Set("max_dedicated_ips", offerSubscription.MaxDedicatedIPs)
	_ = d.Set("max_webhooks_per_domain", offerSubscription.MaxWebhooksPerDomain)
	_ = d.Set("max_custom_blocklists_per_domain", offerSubscription.MaxCustomBlocklistsPerDomain)
	_ = d.Set("included_monthly_emails", offerSubscription.IncludedMonthlyEmails)

	return nil
}
