package webhosting

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceWebhosting() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceWebhosting().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "domain")

	dsSchema["domain"].ConflictsWith = []string{"webhosting_id"}
	dsSchema["webhosting_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Webhosting",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"domain"},
	}
	dsSchema["organization_id"] = account.OrganizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The project ID the resource is associated to",
		ValidateFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceWebhostingRead,
	}
}

func DataSourceWebhostingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	webhostingID, ok := d.GetOk("webhosting_id")
	if !ok {
		hostingDomain := d.Get("domain").(string)
		res, err := api.ListHostings(&webhosting.ListHostingsRequest{
			Region:         region,
			Domain:         types.ExpandStringPtr(hostingDomain),
			ProjectID:      types.ExpandStringPtr(d.Get("project_id")),
			OrganizationID: types.ExpandStringPtr(d.Get("organization_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundDomain, err := datasource.FindExact(
			res.Hostings,
			func(s *webhosting.Hosting) bool { return s.Domain == hostingDomain },
			hostingDomain,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		webhostingID = foundDomain.ID
	}

	regionalID := datasource.NewRegionalID(webhostingID, region)
	d.SetId(regionalID)
	err = d.Set("webhosting_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceWebhostingRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read hosting")...)
	}

	if d.Id() == "" {
		return diag.Errorf("hosting (%s) not found", regionalID)
	}

	return nil
}
