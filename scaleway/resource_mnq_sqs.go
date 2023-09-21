package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayMNQSQS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayMNQSQSCreate,
		ReadContext:   resourceScalewayMNQSQSRead,
		DeleteContext: resourceScalewayMNQSQSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint of the SQS service",
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),
		},
		/* Goal here is to ForceNew if Status is found to be disabled, maybe just add a ForceNew `status` field in schema
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			api, region, id, err := mnqSQSAPIWithRegionAndID(meta, diff.Id())
			if err != nil {
				return err
			}

			sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
				Region:    region,
				ProjectID: id,
			})
			if err != nil {
				return err
			}

			if sqs.Status == mnq.SqsInfoStatusDisabled {
				err := diff.ForceNew("")
				if err != nil {
					return err
				}
			}

			return nil
		},*/
	}
}

func resourceScalewayMNQSQSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQSQSAPI(d, meta)
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

	d.SetId(newRegionalIDString(region, sqs.ProjectID))

	return resourceScalewayMNQSQSRead(ctx, d, meta)
}

func resourceScalewayMNQSQSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqSQSAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("endpoint", sqs.SqsEndpointURL)
	_ = d.Set("region", sqs.Region)
	_ = d.Set("project_id", sqs.ProjectID)

	return nil
}

func resourceScalewayMNQSQSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := mnqSQSAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if sqs.Status == mnq.SqsInfoStatusDisabled {
		d.SetId("")
		return nil
	}

	_, err = api.DeactivateSqs(&mnq.SqsAPIDeactivateSqsRequest{
		Region:    region,
		ProjectID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
