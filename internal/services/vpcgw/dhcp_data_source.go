package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDHCP() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDHCP().SchemaFunc())

	dsSchema["dhcp_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The ID of the public gateway DHCP configuration",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Schema:             dsSchema,
		ReadContext:        dataSourceVPCPublicGatewayDHCPRead,
		DeprecationMessage: dhcpDeprecationMessage,
	}
}

func dataSourceVPCPublicGatewayDHCPRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_vpc_public_gateway_dhcp data source is no longer supported",
		Detail:   dhcpDeprecationMessage,
	}}
}
