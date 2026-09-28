package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

type GlobalIdentity struct {
	ID types.String `tfsdk:"id"`
}

type RegionalIdentity struct {
	ID     types.String `tfsdk:"id"`
	Region types.String `tfsdk:"region"`
}

type ZonalIdentity struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
}

func DefaultGlobal() identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func DefaultRegional() identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"region": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func DefaultZonal() identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"zone": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

// CompositeRegional builds an identity schema for regional resources whose ID is
// composite (e.g. region/deployment_id/name). partKeys are the non-region parts.
func CompositeRegional(partKeys ...string) identityschema.Schema {
	attrs := map[string]identityschema.Attribute{
		"region": identityschema.StringAttribute{
			RequiredForImport: true,
		},
	}

	for _, key := range partKeys {
		attrs[key] = identityschema.StringAttribute{
			RequiredForImport: true,
		}
	}

	return identityschema.Schema{Attributes: attrs}
}

func SetRegionalIdentity(region scw.Region, id string) RegionalIdentity {
	return RegionalIdentity{
		ID:     types.StringValue(regional.NewIDString(region, id)),
		Region: types.StringValue(region.String()),
	}
}

func SetZonalIdentity(zone scw.Zone, id string) ZonalIdentity {
	return ZonalIdentity{
		ID:   types.StringValue(zonal.NewIDString(zone, id)),
		Zone: types.StringValue(zone.String()),
	}
}

func SetGlobalIdentity(id string) GlobalIdentity {
	return GlobalIdentity{
		ID: types.StringValue(id),
	}
}
