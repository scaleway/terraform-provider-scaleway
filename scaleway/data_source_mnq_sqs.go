package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
)

func dataSourceScalewayMNQSQS() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayMNQSQS().Schema)

	addOptionalFieldsToSchema(dsSchema, "region", "project_id")

	return &schema.Resource{
		ReadContext: dataSourceScalewayMNQSQSRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayMNQSQSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQSQSAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if sqs.Status != mnq.SqsInfoStatusEnabled {
		return diag.FromErr(fmt.Errorf("expected mnq sqs status to be enabled, got: %s", sqs.Status))
	}

	regionID := datasourceNewRegionalID(sqs.ProjectID, region)
	d.SetId(regionID)

	diags := resourceScalewayMNQSQSRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read sqs state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("sqs (%s) not found", regionID)
	}

	return nil
}
