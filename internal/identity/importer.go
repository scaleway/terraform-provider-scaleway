package identity

import (
	"context"
	"fmt"

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
