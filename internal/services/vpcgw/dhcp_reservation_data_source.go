package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDHCPReservation() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDHCPReservation().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "mac_address", "gateway_network_id")

	dsSchema["mac_address"].ConflictsWith = []string{"reservation_id"}
	dsSchema["gateway_network_id"].ConflictsWith = []string{"reservation_id"}
	dsSchema["reservation_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of dhcp entry reservation",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"mac_address", "gateway_network_id"},
	}
	dsSchema["wait_for_dhcp"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Wait the MAC address in dhcp entries",
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "zone")

	return &schema.Resource{
		Schema:             dsSchema,
		ReadContext:        dataSourceDHCPReservationRead,
		DeprecationMessage: dhcpDeprecationMessage,
	}
}

func dataSourceDHCPReservationRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "scaleway_vpc_public_gateway_dhcp_reservation data source is no longer supported",
		Detail:   dhcpDeprecationMessage,
	}}
}
