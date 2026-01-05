package identity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// DefaultRegional should be used as the default identity schema for regional resources.
// For instance if you want an id with the form fr-par/11111111-1111-1111-1111-111111111111
func DefaultRegional() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"id": {
			Type:              schema.TypeString,
			Description:       "The id of the resource (UUID format)",
			RequiredForImport: true,
		},
		"region": DefaultRegionAttribute(),
	})
}

func DefaultRegionAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The region of the resource",
		RequiredForImport: true,
	}
}

func DefaultZoneAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The zone of the resource",
		RequiredForImport: true,
	}
}

// DefaultZonal should be used as the default identity schema for zoned resources.
// For instance if you want an id with the form fr-par-1/11111111-1111-1111-1111-111111111111
func DefaultZonal() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"id": {
			Type:              schema.TypeString,
			Description:       "The id of the resource (UUID format)",
			RequiredForImport: true,
		},
		"zone": DefaultZoneAttribute(),
	})
}

func WrapSchemaMap(m map[string]*schema.Schema) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return m
		},
	}
}

func DefaultProjectIDAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The ID of the project (UUID format)",
		RequiredForImport: true,
	}
}

func DefaultProjectID() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"project_id": DefaultProjectIDAttribute(),
	})
}

func SetZonalIdentity(d *schema.ResourceData, zone scw.Zone, id string) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	err = identity.Set("zone", zone.String())
	if err != nil {
		return err
	}

	err = identity.Set("id", id)
	if err != nil {
		return err
	}

	d.SetId(zonal.NewIDString(zone, id))

	return nil
}

func SetRegionalIdentity(d *schema.ResourceData, region scw.Region, id string) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	err = identity.Set("region", region.String())
	if err != nil {
		return err
	}

	err = identity.Set("id", id)
	if err != nil {
		return err
	}

	d.SetId(regional.NewIDString(region, id))

	return nil
}

func SetFlatIdentity(d *schema.ResourceData, key string, value string) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	err = identity.Set(key, value)
	if err != nil {
		return err
	}

	d.SetId(value)

	return nil
}
