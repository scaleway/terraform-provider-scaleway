package s2svpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2svpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCustomerGateway() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCustomerGateway().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"customer_gateway_id"}
	dsSchema["customer_gateway_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the customer gateway",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceS2SCustomerGatewayRead,
	}
}

func DataSourceS2SCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	customerGatewayID, ok := d.GetOk("customer_gateway_id")
	if !ok {
		gatewayName := d.Get("name").(string)

		res, err := api.ListCustomerGateways(&s2svpn.ListCustomerGatewaysRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(gatewayName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundGateway, err := datasource.FindExact(
			res.Gateways,
			func(s *s2svpn.CustomerGateway) bool { return s.Name == gatewayName },
			gatewayName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		customerGatewayID = foundGateway.ID
	}

	regionalID := datasource.NewRegionalID(customerGatewayID, region)
	d.SetId(regionalID)
	_ = d.Set("customer_gateway_id", regionalID)

	diags := ResourceCustomerGatewayRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read customer gateway state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("customer gateway (%s) not found", regionalID)
	}

	return nil
}
