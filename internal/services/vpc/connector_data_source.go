package vpc

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/connector_data_source.md
var connectorDataSourceDescription string

func DataSourceConnector() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceConnector().SchemaFunc())

	filterFields := []string{"name", "vpc_id", "target_vpc_id", "tags", "project_id"}

	dsSchema["connector_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the VPC connector",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "vpc_id", "target_vpc_id", "tags", "project_id", "region")

	for _, f := range filterFields {
		dsSchema[f].ConflictsWith = []string{"connector_id"}
	}

	return &schema.Resource{
		ReadContext: DataSourceConnectorRead,
		Description: connectorDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	connectorID, idExists := d.GetOk("connector_id")
	if idExists {
		return dataSourceConnectorReadByID(ctx, d, m, connectorID.(string))
	}

	return dataSourceConnectorReadByFilters(ctx, d, m)
}

func dataSourceConnectorReadByID(ctx context.Context, d *schema.ResourceData, m any, connectorID string) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	id := locality.ExpandID(connectorID)
	d.SetId(regional.NewIDString(region, id))

	connector, err := vpcAPI.GetVPCConnector(&vpc.GetVPCConnectorRequest{
		Region:         region,
		VpcConnectorID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setConnectorState(d, connector)
}

func dataSourceConnectorReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.ListVPCConnectorsRequest{
		Region:    region,
		Name:      types.ExpandStringPtr(d.Get("name")),
		Tags:      types.ExpandStrings(d.Get("tags")),
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
	}

	if vpcID, ok := d.GetOk("vpc_id"); ok {
		req.VpcID = new(locality.ExpandID(vpcID.(string)))
	}

	if targetVpcID, ok := d.GetOk("target_vpc_id"); ok {
		req.TargetVpcID = new(locality.ExpandID(targetVpcID.(string)))
	}

	res, err := vpcAPI.ListVPCConnectors(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.VpcConnectors) == 0 {
		return diag.FromErr(errors.New("no VPC connector found matching the specified filters"))
	}

	if len(res.VpcConnectors) > 1 {
		return diag.FromErr(fmt.Errorf("multiple VPC connectors (%d) found, please refine your filters or use connector_id", len(res.VpcConnectors)))
	}

	connector := res.VpcConnectors[0]
	d.SetId(regional.NewIDString(connector.Region, connector.ID))

	return setConnectorState(d, connector)
}
