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

func DataSourceConnection() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceConnection().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"connection_id"}
	dsSchema["connection_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the connection",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceS2SConnectionRead,
	}
}

func DataSourceS2SConnectionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	connectionID, ok := d.GetOk("connection_id")
	if !ok {
		connectionName := d.Get("name").(string)

		res, err := api.ListConnections(&s2svpn.ListConnectionsRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(connectionName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundConnection, err := datasource.FindExact(
			res.Connections,
			func(s *s2svpn.Connection) bool { return s.Name == connectionName },
			connectionName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		connectionID = foundConnection.ID
	}

	regionalID := datasource.NewRegionalID(connectionID, region)
	d.SetId(regionalID)
	_ = d.Set("connection_id", regionalID)

	diags := ResourceConnectionRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read connection state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("connection (%s) not found", regionalID)
	}

	return nil
}
