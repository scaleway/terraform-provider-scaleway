package mnq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceSQS() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSQS().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "region", "project_id")

	return &schema.Resource{
		ReadContext: DataSourceMNQSQSRead,
		Schema:      dsSchema,
	}
}

func DataSourceMNQSQSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newSQSAPI(d, m)
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

	regionID := datasource.NewRegionalID(sqs.ProjectID, region)
	d.SetId(regionID)

	diags := ResourceMNQSQSRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read sqs state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("sqs (%s) not found", regionID)
	}

	return nil
}
