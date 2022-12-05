package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayBaremetalPrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayBaremetalPrivateNetworkCreate,
		ReadContext:   resourceScalewayBaremetalPrivateNetworkRead,
		UpdateContext: resourceScalewayBaremetalPrivateNetworkUpdate,
		DeleteContext: resourceScalewayBaremetalPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Create:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Update:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
			Delete:  schema.DefaultTimeout(defaultBaremetalServerTimeout),
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:             schema.TypeString,
				Description:      "The server ID",
				Required:         true,
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"private_networks": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The private network ID",
							Required:    true,
						},
						"project_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private network project ID",
						},
						"vlan": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The VLAN ID associated to the private network",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private network status",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the creation of the private network",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the last update of the private network",
						},
					},
				},
				Required:    true,
				Description: "The private networks to attach to the server",
			},
			"zone": zoneSchema(),
		},
	}
}

func resourceScalewayBaremetalPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalPrivateNetworkAPI, zone, err := baremetalPrivateNetworkAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworkIDs := d.Get("private_networks")

	createBaremetalPrivateNetworkRequest := &baremetal.PrivateNetworkAPISetServerPrivateNetworksRequest{
		Zone:              zone,
		ServerID:          expandZonedID(d.Get("server_id").(string)).ID,
		PrivateNetworkIDs: expandBaremetalPrivateNetworks(privateNetworkIDs),
	}

	baremetalPrivateNetwork, err := baremetalPrivateNetworkAPI.SetServerPrivateNetworks(
		createBaremetalPrivateNetworkRequest,
		scw.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBaremetalServerPrivateNetwork(ctx, baremetalPrivateNetworkAPI, zone, baremetalPrivateNetwork.ServerPrivateNetworks[0].ServerID, d.Timeout(schema.TimeoutCreate))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	d.SetId(
		newZonedNestedIDString(
			zone,
			baremetalPrivateNetwork.ServerPrivateNetworks[0].ServerID,
			"",
		),
	)

	return resourceScalewayBaremetalPrivateNetworkRead(ctx, d, meta)
}

func resourceScalewayBaremetalPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalPrivateNetworkAPI, _, err := baremetalPrivateNetworkAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, _, serverID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	listPrivateNetworks, err := baremetalPrivateNetworkAPI.ListServerPrivateNetworks(&baremetal.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     zone,
		ServerID: &serverID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list server's private networks: %w", err))
	}

	_ = d.Set("zone", zone)
	_ = d.Set("server_id", newZonedID(zone, serverID).String())
	_ = d.Set("private_networks", flattenBaremetalPrivateNetworks(zone, listPrivateNetworks.ServerPrivateNetworks))

	return nil
}

func resourceScalewayBaremetalPrivateNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalPrivateNetworkAPI, zone, err := baremetalPrivateNetworkAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("private_networks", "server_id") {
		privateNetworkIDs := d.Get("private_networks")

		updateBaremetalPrivateNetworkRequest := &baremetal.PrivateNetworkAPISetServerPrivateNetworksRequest{
			Zone:              zone,
			ServerID:          expandZonedID(d.Get("server_id").(string)).ID,
			PrivateNetworkIDs: expandBaremetalPrivateNetworks(privateNetworkIDs),
		}

		baremetalPrivateNetwork, err := baremetalPrivateNetworkAPI.SetServerPrivateNetworks(
			updateBaremetalPrivateNetworkRequest,
			scw.WithContext(ctx),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForBaremetalServerPrivateNetwork(ctx, baremetalPrivateNetworkAPI, zone, baremetalPrivateNetwork.ServerPrivateNetworks[0].ServerID, d.Timeout(schema.TimeoutUpdate))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}

		d.SetId(
			newZonedNestedIDString(
				zone,
				baremetalPrivateNetwork.ServerPrivateNetworks[0].ServerID,
				"",
			),
		)
	}

	return resourceScalewayBaremetalPrivateNetworkRead(ctx, d, meta)
}

func resourceScalewayBaremetalPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalPrivateNetworkAPI, _, err := baremetalPrivateNetworkAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, _, serverID, err := parseZonedNestedID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	listPrivateNetworks, err := baremetalPrivateNetworkAPI.ListServerPrivateNetworks(&baremetal.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     zone,
		ServerID: &serverID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list server's private networks: %w", err))
	}

	var privateNetworkIDs []string
	for _, privateNetworkID := range listPrivateNetworks.ServerPrivateNetworks {
		privateNetworkIDs = append(privateNetworkIDs, newZonedID(zone, privateNetworkID.PrivateNetworkID).ID)
	}

	for i := range privateNetworkIDs {
		err = baremetalPrivateNetworkAPI.DeleteServerPrivateNetwork(&baremetal.PrivateNetworkAPIDeleteServerPrivateNetworkRequest{
			Zone:             zone,
			ServerID:         serverID,
			PrivateNetworkID: privateNetworkIDs[i],
		}, scw.WithContext(ctx))
		if err != nil && !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForBaremetalServerPrivateNetwork(ctx, baremetalPrivateNetworkAPI, zone, serverID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
