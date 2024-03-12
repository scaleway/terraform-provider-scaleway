package scaleway

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayInstancePrivateNIC() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstancePrivateNIC().Schema)

	addOptionalFieldsToSchema(dsSchema, "private_network_id", "zone", "tags")
	fixDatasourceSchemaFlags(dsSchema, true, "server_id")

	dsSchema["private_network_id"].ConflictsWith = []string{"private_nic_id"}

	dsSchema["private_nic_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Private NIC",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"private_network_id"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstancePrivateNICRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayInstancePrivateNICRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	serverID := locality.ExpandID(d.Get("server_id"))

	id, ok := d.GetOk("private_nic_id")
	var privateNICID string
	if !ok {
		resp, err := instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{
			Zone:     zone,
			ServerID: serverID,
			Tags:     expandStrings(d.Get("tags")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to list instance private_nic: %w", err))
		}

		privateNic, err := privateNICWithFilters(resp.PrivateNics, d)
		if err != nil {
			return diag.FromErr(err)
		}

		privateNICID = privateNic.ID
	} else {
		_, privateNICID, _ = locality.ParseLocalizedID(id.(string))
	}

	zonedID := zonal.NewNestedIDString(
		zone,
		serverID,
		privateNICID,
	)
	d.SetId(zonedID)
	err = d.Set("private_nic_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayInstancePrivateNICRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read private nic state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("instance private nic (%s) not found", zonedID)
	}

	return nil
}

func privateNICWithFilters(privateNICs []*instance.PrivateNIC, d *schema.ResourceData) (*instance.PrivateNIC, error) {
	privateNetworkID := locality.ExpandID(d.Get("private_network_id"))

	if privateNetworkID == "" {
		switch {
		case len(privateNICs) == 1:
			return privateNICs[0], nil
		case len(privateNICs) == 0:
			return nil, errors.New("found no private nic with given filters")
		default:
			return nil, errors.New("found more than one private nic with given filters")
		}
	}

	var privateNIC *instance.PrivateNIC

	for _, pnic := range privateNICs {
		if pnic.PrivateNetworkID == privateNetworkID {
			if privateNIC != nil {
				return nil, fmt.Errorf("found more than one private nic for request private network (%s)", privateNetworkID)
			}
			privateNIC = pnic
		}
	}

	if privateNIC != nil {
		return privateNIC, nil
	}

	return nil, fmt.Errorf("could not find a private_nic for private network (%s)", privateNetworkID)
}
