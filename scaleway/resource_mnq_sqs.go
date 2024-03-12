package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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
			"region":     regional.Schema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayMNQSQSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newMNQSQSAPI(d, m.(*meta.Meta))
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

	d.SetId(regional.NewIDString(region, sqs.ProjectID))

	return resourceScalewayMNQSQSRead(ctx, d, m)
}

func resourceScalewayMNQSQSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := mnqSQSAPIWithRegionAndID(m.(*meta.Meta), d.Id())
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

func resourceScalewayMNQSQSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := mnqSQSAPIWithRegionAndID(m.(*meta.Meta), d.Id())
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
