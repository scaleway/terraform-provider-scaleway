package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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

func dataSourceScalewayMNQSQSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newMNQSQSAPI(d, m.(*meta.Meta))
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

	diags := resourceScalewayMNQSQSRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read sqs state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("sqs (%s) not found", regionID)
	}

	return nil
}
