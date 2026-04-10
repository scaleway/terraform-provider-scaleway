package mnq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceSNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQSNSCreate,
		ReadContext:   ResourceMNQSNSRead,
		DeleteContext: ResourceMNQSNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    snsSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func snsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoint": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Endpoint of the SNS service",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceMNQSNSCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newMNQSNSAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	sns, err := api.ActivateSns(&mnq.SnsAPIActivateSnsRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, sns.Region, sns.ProjectID); err != nil {
		return diag.FromErr(err)
	}

	return ResourceMNQSNSRead(ctx, d, m)
}

func ResourceMNQSNSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSNSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sns, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sns, err = api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
				Region:    region,
				ProjectID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, region, id); err != nil {
		return diag.FromErr(err)
	}

	return setSNSState(d, sns)
}

// readSNSIntoState fetches the SNS info and sets state without calling identity.SetRegionalIdentity.
// Use this for data sources which do not have Identity schema.
func readSNSIntoState(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSNSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sns, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sns, err = api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
				Region:    region,
				ProjectID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return setSNSState(d, sns)
}

func setSNSState(d *schema.ResourceData, sns *mnq.SnsInfo) diag.Diagnostics {
	_ = d.Set("endpoint", sns.SnsEndpointURL)
	_ = d.Set("region", sns.Region)
	_ = d.Set("project_id", sns.ProjectID)

	return nil
}

func ResourceMNQSNSDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSNSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sns, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sns, err = api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
				Region:    region,
				ProjectID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if sns.Status == mnq.SnsInfoStatusDisabled {
		d.SetId("")

		return nil
	}

	err = retryMNQNamespaceRead(ctx, func() error {
		_, e := api.DeactivateSns(&mnq.SnsAPIDeactivateSnsRequest{
			Region:    region,
			ProjectID: id,
		}, scw.WithContext(ctx))

		return e
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
