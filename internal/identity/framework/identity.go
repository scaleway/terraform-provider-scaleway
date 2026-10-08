package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
		ID:     types.StringValue(id),
		Region: types.StringValue(region.String()),
	}
}

func SetZonalIdentity(zone scw.Zone, id string) ZonalIdentity {
	return ZonalIdentity{
		ID:   types.StringValue(id),
		Zone: types.StringValue(zone.String()),
	}
}

func SetGlobalIdentity(id string) GlobalIdentity {
	return GlobalIdentity{
		ID: types.StringValue(id),
	}
}

// ResolveZonalIdentity returns the zone and bare UUID from a zonal identity,
// tolerating identity state written by older provider versions where id was
// stored as "zone/uuid" instead of a bare UUID. State self-heals on the next
// Identity.Set, which writes the bare-uuid format.
func ResolveZonalIdentity(identity ZonalIdentity) (scw.Zone, string) {
	id := identity.ID.ValueString()

	if zone, parsedID, err := zonal.ParseID(id); err == nil {
		return zone, parsedID
	}

	return scw.Zone(identity.Zone.ValueString()), id
}

// ResolveRegionalIdentity returns the region and bare UUID from a regional
// identity, tolerating identity state written by older provider versions where
// id was stored as "region/uuid" instead of a bare UUID. State self-heals on
// the next Identity.Set, which writes the bare-uuid format.
func ResolveRegionalIdentity(identity RegionalIdentity) (scw.Region, string) {
	id := identity.ID.ValueString()

	if region, parsedID, err := regional.ParseID(id); err == nil {
		return region, parsedID
	}

	return scw.Region(identity.Region.ValueString()), id
}
