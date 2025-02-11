package flexibleip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceFlexibleIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceIP().Schema)

	dsSchema["ip_address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The IPv4 address",
		ConflictsWith: []string{"flexible_ip_id"},
	}
	dsSchema["flexible_ip_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IPv4 address",
		ConflictsWith:    []string{"ip_address"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}
	dsSchema["project_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Description:      "The project_id you want to attach the resource to",
		Optional:         true,
		ForceNew:         true,
		Computed:         true,
		ValidateDiagFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: DataSourceFlexibleIPRead,
		Schema:      dsSchema,
	}
}

func DataSourceFlexibleIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, ipIDExists := d.GetOk("flexible_ip_id")

	if !ipIDExists {
		res, err := fipAPI.ListFlexibleIPs(&flexibleip.ListFlexibleIPsRequest{
			Zone:      zone,
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, ip := range res.FlexibleIPs {
			if ip.IPAddress.String() == d.Get("ip_address").(string) {
				if ipID != "" {
					return diag.Errorf("more than 1 flexible ip found with the same IPv4 address %s", d.Get("ip_address"))
				}

				ipID = ip.ID
			}
		}

		if ipID == "" {
			return diag.Errorf("no flexible ip found with the same IPv4 address %s", d.Get("ip_address"))
		}
	}

	zoneID := datasource.NewZonedID(ipID, zone)
	d.SetId(zoneID)

	err = d.Set("flexible_ip_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceFlexibleIPRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read flexible ip state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("flexible ip (%s) not found", ipID)
	}

	return nil
}
