package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstancePrivateNIC() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstancePrivateNIC().Schema)

	addOptionalFieldsToSchema(dsSchema, "private_network_id", "zone")
	fixDatasourceSchemaFlags(dsSchema, true, "server_id")

	dsSchema["private_network_id"].ConflictsWith = []string{"private_nic_id"}
	dsSchema["private_network_id"].AtLeastOneOf = []string{"private_nic_id"}

	dsSchema["private_nic_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Private NIC",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		AtLeastOneOf:  []string{"private_network_id"},
		ConflictsWith: []string{"private_network_id"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstancePrivateNICRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayInstancePrivateNICRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID := expandID(d.Get("server_id"))

	id, ok := d.GetOk("private_nic_id")
	var privateNICID string
	if !ok {
		privateNetworkID := expandID(d.Get("private_network_id"))
		privateNic, err := privateNICWithPrivateNetworkID(ctx, instanceAPI, zone, serverID, privateNetworkID)
		if err != nil {
			return diag.FromErr(err)
		}
		privateNICID = privateNic.ID
	} else {
		_, privateNICID, _ = parseLocalizedID(id.(string))
	}

	zonedID := newZonedNestedIDString(
		zone,
		serverID,
		privateNICID,
	)
	d.SetId(zonedID)
	err = d.Set("private_nic_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayInstancePrivateNICRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read private nic state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("instance private nic (%s) not found", zonedID)
	}

	return nil
}

func privateNICWithPrivateNetworkID(ctx context.Context, api *instance.API, zone scw.Zone, serverID, privateNetworkID string) (*instance.PrivateNIC, error) {
	resp, err := api.ListPrivateNICs(&instance.ListPrivateNICsRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list instance private_nic: %w", err)
	}
	for _, pnic := range resp.PrivateNics {
		if pnic.PrivateNetworkID == privateNetworkID {
			return pnic, nil
		}
	}
	return nil, fmt.Errorf("could not find a private_nic for server (%s) and private network (%s) in zone (%s)", serverID, privateNetworkID, zone)
}
