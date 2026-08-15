package identity

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func DefaultRegionalImporter() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
			// If importing by ID, we just set the ID field to state, allowing the read to fill in the rest of the data.
			if d.Id() != "" {
				return []*schema.ResourceData{d}, nil
			}

			importedIdentity, err := d.Identity()
			if err != nil {
				return nil, fmt.Errorf("error getting identity: %w", err)
			}

			region := importedIdentity.Get("region").(string)
			id := importedIdentity.Get("id").(string)

			err = SetRegionalIdentity(d, scw.Region(region), id)
			if err != nil {
				return nil, err
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}

// CompositeRegionalImporter creates an importer for regional resources with
// composite IDs (e.g., region/instanceID/databaseName). It handles both
// ID-based and identity-based imports. For identity-based imports, it
// constructs d.Id() by joining the identity attributes in keyOrder with "/".
// The keyOrder must include "region" as the first element.
func CompositeRegionalImporter(keyOrder ...string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
			// If importing by ID, we just set the ID field to state, allowing the read to fill in the rest of the data.
			if d.Id() != "" {
				return []*schema.ResourceData{d}, nil
			}

			importedIdentity, err := d.Identity()
			if err != nil {
				return nil, fmt.Errorf("error getting identity: %w", err)
			}

			parts := make([]string, len(keyOrder))
			for i, key := range keyOrder {
				val, ok := importedIdentity.GetOk(key)
				if !ok {
					return nil, fmt.Errorf("expected identity to contain key %s", key)
				}

				str, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf("expected identity key %s to be a string, was: %T", key, val)
				}

				parts[i] = str
			}

			d.SetId(strings.Join(parts, "/"))

			return []*schema.ResourceData{d}, nil
		},
	}
}
