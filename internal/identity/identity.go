package identity

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

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

// DefaultGlobal should be used as the default identity schema for global/flat resources.
// For instance if you want an id with the form 11111111-1111-1111-1111-111111111111 (UUID only)
func DefaultGlobal() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"id": {
			Type:              schema.TypeString,
			Description:       "The id of the resource (UUID format)",
			RequiredForImport: true,
		},
	})
}

func WrapSchemaMap(m map[string]*schema.Schema) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return m
		},
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

func SetGlobalIdentity(d *schema.ResourceData, id string) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	err = identity.Set("id", id)
	if err != nil {
		return err
	}

	d.SetId(id)

	return nil
}

// SetMultiPartIdentity sets identity attributes and constructs a composite ID from multiple parts
func SetMultiPartIdentity(d *schema.ResourceData, values map[string]string, keyOrder ...string) error {
	if len(keyOrder) != len(values) {
		return fmt.Errorf("keyOrder length (%d) does not match values length (%d)", len(keyOrder), len(values))
	}

	for _, key := range keyOrder {
		if _, exists := values[key]; !exists {
			return fmt.Errorf("key %q from keyOrder not found in values", key)
		}
	}

	identity, err := d.Identity()
	if err != nil {
		return err
	}

	for key, value := range values {
		err = identity.Set(key, value)
		if err != nil {
			return err
		}
	}

	parts := make([]string, len(keyOrder))
	for i, key := range keyOrder {
		parts[i] = values[key]
	}

	d.SetId(strings.Join(parts, "/"))

	return nil
}

// ParseMultiPartID extracts identity parts from a composite ID
func ParseMultiPartID(id string, keyOrder ...string) map[string]string {
	parts := strings.SplitN(id, "/", len(keyOrder))
	result := make(map[string]string, len(keyOrder))

	for i, key := range keyOrder {
		if i < len(parts) {
			result[key] = parts[i]
		}
	}

	return result
}
