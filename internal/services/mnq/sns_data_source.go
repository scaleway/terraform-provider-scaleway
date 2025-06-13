package mnq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceSNS() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSNS().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "region", "project_id")

	return &schema.Resource{
		ReadContext: DataSourceMNQSNSRead,
		Schema:      dsSchema,
	}
}

func DataSourceMNQSNSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newMNQSNSAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	sns, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if sns.Status != mnq.SnsInfoStatusEnabled {
		return diag.FromErr(fmt.Errorf("expected mnq sns status to be enabled, got: %s", sns.Status))
	}

	regionID := datasource.NewRegionalID(sns.ProjectID, region)
	d.SetId(regionID)

	diags := ResourceMNQSNSRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read sns state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("sns (%s) not found", regionID)
	}

	return nil
}
