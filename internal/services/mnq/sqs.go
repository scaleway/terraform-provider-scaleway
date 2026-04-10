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

func ResourceSQS() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceMNQSQSCreate,
		ReadContext:   ResourceMNQSQSRead,
		DeleteContext: ResourceMNQSQSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    sqsSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func sqsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoint": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Endpoint of the SQS service",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceMNQSQSCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newSQSAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.ActivateSqs(&mnq.SqsAPIActivateSqsRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.SetRegionalIdentity(d, region, sqs.ProjectID); err != nil {
		return diag.FromErr(err)
	}

	return ResourceMNQSQSRead(ctx, d, m)
}

func ResourceMNQSQSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sqs, err = api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
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

	return setSQSState(d, sqs)
}

// readSQSIntoState fetches the SQS info and sets state without calling identity.SetRegionalIdentity.
// Use this for data sources which do not have Identity schema.
func readSQSIntoState(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sqs, err = api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
				Region:    region,
				ProjectID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return setSQSState(d, sqs)
}

func setSQSState(d *schema.ResourceData, sqs *mnq.SqsInfo) diag.Diagnostics {
	_ = d.Set("endpoint", sqs.SqsEndpointURL)
	_ = d.Set("region", sqs.Region)
	_ = d.Set("project_id", sqs.ProjectID)

	return nil
}

func ResourceMNQSQSDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewSQSAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil && isMNQNamespaceReadRetryableError(err) {
		err = retryMNQNamespaceRead(ctx, func() error {
			sqs, err = api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
				Region:    region,
				ProjectID: id,
			}, scw.WithContext(ctx))

			return err
		})
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if sqs.Status == mnq.SqsInfoStatusDisabled {
		d.SetId("")

		return nil
	}

	err = retryMNQNamespaceRead(ctx, func() error {
		_, e := api.DeactivateSqs(&mnq.SqsAPIDeactivateSqsRequest{
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
