package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instance "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourcePrivateNIC() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePrivateNIC().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "private_network_id", "zone", "tags")
	datasource.FixDatasourceSchemaFlags(dsSchema, true, "server_id")

	dsSchema["private_network_id"].ConflictsWith = []string{"private_nic_id"}

	dsSchema["private_nic_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the Private NIC",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"private_network_id"},
	}

	return &schema.Resource{
		ReadContext: DataSourceInstancePrivateNICRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstancePrivateNICRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIV2WithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var privateNICID string

	var pNIC *instance.PrivateNetworkInterface

	serverID := locality.ExpandID(d.Get("server_id"))

	id, ok := d.GetOk("private_nic_id")
	if !ok {
		resp, err := instanceAPI.ListPrivateNetworkInterfaces(&instance.ListPrivateNetworkInterfacesRequest{
			Zone:      zone,
			ServerIDs: []string{serverID},
			Tags:      types.ExpandStrings(d.Get("tags")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to list instance private_nic: %w", err))
		}

		privateNic, err := privateNICWithFilters(resp.PrivateNetworkInterfaces, d)
		if err != nil {
			return diag.FromErr(err)
		}

		pNIC = privateNic
		privateNICID = privateNic.ID
	} else {
		pNICID, err := locality.ExtractUUID(id.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		privateNetworkInterface, err := instanceAPI.GetPrivateNetworkInterface(&instance.GetPrivateNetworkInterfaceRequest{
			Zone:                      zone,
			PrivateNetworkInterfaceID: pNICID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		pNIC = privateNetworkInterface
		privateNICID = privateNetworkInterface.ID
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

	instanceAPIV1, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPrivateNICState(ctx, instanceAPIV1, d, pNIC, zone, m)
}

func privateNICWithFilters(privateNICs []*instance.PrivateNetworkInterface, d *schema.ResourceData) (*instance.PrivateNetworkInterface, error) {
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

	var privateNIC *instance.PrivateNetworkInterface

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
