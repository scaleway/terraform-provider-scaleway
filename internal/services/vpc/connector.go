package vpc

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/connector_resource.md
var connectorResourceDescription string

func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceConnectorCreate,
		ReadContext:   ResourceConnectorRead,
		UpdateContext: ResourceConnectorUpdate,
		DeleteContext: ResourceConnectorDelete,
		Description:   connectorResourceDescription,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    connectorSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func connectorSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the VPC connector",
			Computed:    true,
		},
		"tags": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The tags associated with the VPC connector",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"vpc_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the source VPC",
		},
		"target_vpc_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The ID of the target VPC to connect to",
		},
		"project_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Scaleway Project the VPC connector belongs to",
		},
		"region": regional.Schema(),
		// Computed elements
		"organization_id": account.OrganizationIDSchema(),
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The VPC connector status",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the vpc connector",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the vpc connector",
		},
	}
}

func ResourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.CreateVPCConnector(&vpc.CreateVPCConnectorRequest{
		Name:        types.ExpandOrGenerateString(d.Get("name"), "connector"),
		VpcID:       regional.ExpandID(d.Get("vpc_id").(string)).ID,
		TargetVpcID: regional.ExpandID(d.Get("target_vpc_id").(string)).ID,
		Tags:        types.ExpandStrings(d.Get("tags")),
		Region:      region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, res.Region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceConnectorRead(ctx, d, m)
}

func ResourceConnectorRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.GetVPCConnector(&vpc.GetVPCConnectorRequest{
		Region:         region,
		VpcConnectorID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setConnectorState(d, res)

	err = identity.SetRegionalIdentity(d, res.Region, ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setConnectorState(d *schema.ResourceData, connector *vpc.VPCConnector) diag.Diagnostics {
	_ = d.Set("name", connector.Name)
	_ = d.Set("vpc_id", regional.NewIDString(connector.Region, connector.VpcID))
	_ = d.Set("target_vpc_id", regional.NewIDString(connector.Region, connector.TargetVpcID))
	_ = d.Set("organization_id", connector.OrganizationID)
	_ = d.Set("project_id", connector.ProjectID)
	_ = d.Set("created_at", types.FlattenTime(connector.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(connector.UpdatedAt))
	_ = d.Set("status", connector.Status.String())
	_ = d.Set("region", connector.Region)
	_ = d.Set("tags", connector.Tags)

	return nil
}

func ResourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	updateRequest := &vpc.UpdateVPCConnectorRequest{
		Region:         region,
		VpcConnectorID: ID,
	}

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = vpcAPI.UpdateVPCConnector(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceConnectorRead(ctx, d, m)
}

func ResourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcAPI.DeleteVPCConnector(&vpc.DeleteVPCConnectorRequest{
		Region:         region,
		VpcConnectorID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
